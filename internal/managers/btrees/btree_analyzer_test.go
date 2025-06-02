package btrees

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewBTreeAnalyzer(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	if analyzer == nil {
		t.Fatal("NewBTreeAnalyzer returned nil")
	}
}

func TestBTreeAnalyzer_GetNodeDistribution(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	levels, err := analyzer.GetNodeDistribution()
	if err != nil {
		t.Fatalf("GetNodeDistribution failed: %v", err)
	}

	if len(levels) != 2 {
		t.Errorf("Expected 2 levels, got %d", len(levels))
	}

	// Check level 0 (root)
	if len(levels) > 0 {
		level0 := levels[0]
		if level0.Level != 0 {
			t.Errorf("Level 0: expected level 0, got %d", level0.Level)
		}
		if level0.NodeCount != 1 {
			t.Errorf("Level 0: expected 1 node, got %d", level0.NodeCount)
		}
		if level0.AverageKeyCount != 3.0 {
			t.Errorf("Level 0: expected average key count 3.0, got %f", level0.AverageKeyCount)
		}
		if level0.MinKeyCount != 3 || level0.MaxKeyCount != 3 {
			t.Errorf("Level 0: expected min/max key count 3/3, got %d/%d", level0.MinKeyCount, level0.MaxKeyCount)
		}
	}

	// Check level 1 (leaves)
	if len(levels) > 1 {
		level1 := levels[1]
		if level1.Level != 1 {
			t.Errorf("Level 1: expected level 1, got %d", level1.Level)
		}
		if level1.NodeCount != 4 {
			t.Errorf("Level 1: expected 4 nodes, got %d", level1.NodeCount)
		}
		if level1.AverageKeyCount != 3.0 {
			t.Errorf("Level 1: expected average key count 3.0, got %f", level1.AverageKeyCount)
		}
		if level1.MinKeyCount != 3 || level1.MaxKeyCount != 3 {
			t.Errorf("Level 1: expected min/max key count 3/3, got %d/%d", level1.MinKeyCount, level1.MaxKeyCount)
		}
	}
}

func TestBTreeAnalyzer_CalculateHeight(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	height, err := analyzer.CalculateHeight()
	if err != nil {
		t.Fatalf("CalculateHeight failed: %v", err)
	}

	// Our test tree has 2 levels: root (level 1) + leaves (level 0), so height = 2
	if height != 2 {
		t.Errorf("Expected height 2, got %d", height)
	}
}

func TestBTreeAnalyzer_CalculateFillFactor(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	fillFactor, err := analyzer.CalculateFillFactor()
	if err != nil {
		t.Fatalf("CalculateFillFactor failed: %v", err)
	}

	// Fill factor should be between 0 and 100
	if fillFactor < 0 || fillFactor > 100 {
		t.Errorf("Fill factor %f should be between 0 and 100", fillFactor)
	}

	// For our test tree, fill factor should be relatively low since we have small nodes
	if fillFactor > 50 {
		t.Errorf("Fill factor %f seems too high for test tree", fillFactor)
	}
}

func TestBTreeAnalyzer_AnalyzeStructure(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	analysis, err := analyzer.AnalyzeStructure()
	if err != nil {
		t.Fatalf("AnalyzeStructure failed: %v", err)
	}

	// Check basic structure
	if analysis.Height != 2 {
		t.Errorf("Expected height 2, got %d", analysis.Height)
	}

	if analysis.TotalNodes != 5 { // 1 root + 4 leaves
		t.Errorf("Expected 5 total nodes, got %d", analysis.TotalNodes)
	}

	if analysis.TotalKeys != 15 { // 3 keys per node * 5 nodes
		t.Errorf("Expected 15 total keys, got %d", analysis.TotalKeys)
	}

	if analysis.FillFactor < 0 || analysis.FillFactor > 100 {
		t.Errorf("Fill factor %f should be between 0 and 100", analysis.FillFactor)
	}

	if len(analysis.Levels) != 2 {
		t.Errorf("Expected 2 levels, got %d", len(analysis.Levels))
	}

	if !analysis.IsBalanced {
		t.Error("Tree should be balanced")
	}

	// Check extreme nodes
	if analysis.LargestNode.KeyCount != 3 {
		t.Errorf("Expected largest node key count 3, got %d", analysis.LargestNode.KeyCount)
	}

	if analysis.SmallestNode.KeyCount != 3 {
		t.Errorf("Expected smallest node key count 3, got %d", analysis.SmallestNode.KeyCount)
	}
}

func TestBTreeAnalyzer_AdditionalAnalysis(t *testing.T) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo).(*btreeAnalyzer)

	// Test AnalyzeKeyDistribution
	keyDist, err := analyzer.AnalyzeKeyDistribution()
	if err != nil {
		t.Fatalf("AnalyzeKeyDistribution failed: %v", err)
	}

	if keyDist.TotalKeys != 15 { // 3 keys * 5 nodes (changed from 12 to 15)
		t.Errorf("Expected 15 total keys, got %d", keyDist.TotalKeys)
	}

	if keyDist.LeafKeys != 12 { // 3 keys * 4 leaf nodes (changed from 9 to 12)
		t.Errorf("Expected 12 leaf keys, got %d", keyDist.LeafKeys)
	}

	if keyDist.InternalKeys != 3 { // 3 keys * 1 internal node
		t.Errorf("Expected 3 internal keys, got %d", keyDist.InternalKeys)
	}

	if keyDist.MinKeysPerNode != 3 {
		t.Errorf("Expected min keys per node 3, got %d", keyDist.MinKeysPerNode)
	}

	if keyDist.MaxKeysPerNode != 3 {
		t.Errorf("Expected max keys per node 3, got %d", keyDist.MaxKeysPerNode)
	}

	if keyDist.AvgKeysPerNode != 3.0 {
		t.Errorf("Expected average keys per node 3.0, got %f", keyDist.AvgKeysPerNode)
	}

	if keyDist.LeafNodeCount != 4 { // Changed from 3 to 4
		t.Errorf("Expected 4 leaf nodes, got %d", keyDist.LeafNodeCount)
	}

	if keyDist.InternalNodeCount != 1 {
		t.Errorf("Expected 1 internal node, got %d", keyDist.InternalNodeCount)
	}

	// Test AnalyzeStorageEfficiency
	storageEff, err := analyzer.AnalyzeStorageEfficiency()
	if err != nil {
		t.Fatalf("AnalyzeStorageEfficiency failed: %v", err)
	}

	if storageEff.NodeCount != 5 { // Changed from 4 to 5
		t.Errorf("Expected 5 nodes, got %d", storageEff.NodeCount)
	}

	if storageEff.TotalCapacity == 0 {
		t.Error("Total capacity should not be zero")
	}

	if storageEff.UsedSpace == 0 {
		t.Error("Used space should not be zero")
	}

	if storageEff.Utilization < 0 || storageEff.Utilization > 100 {
		t.Errorf("Utilization %f should be between 0 and 100", storageEff.Utilization)
	}

	if storageEff.AvgNodeSize <= 0 {
		t.Errorf("Average node size %f should be positive", storageEff.AvgNodeSize)
	}
}

func TestBTreeAnalyzer_ErrorCases(t *testing.T) {
	// Test with invalid node size
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	// Set invalid node size
	btreeInfo.(*MockBTreeInfoReader).nodeSize = 0

	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	_, err := analyzer.CalculateFillFactor()
	if err == nil {
		t.Error("Expected error when node size is 0")
	}

	// Test with missing root
	blockReader := NewMockBlockDeviceReader()
	btreeInfo2 := NewMockBTreeInfoReader()
	invalidNavigator := NewBTreeNavigator(blockReader, 9999, btreeInfo2) // Non-existent root
	invalidTraverser := NewBTreeTraverser(invalidNavigator)
	invalidAnalyzer := NewBTreeAnalyzer(invalidNavigator, invalidTraverser, btreeInfo2)

	_, err = invalidAnalyzer.AnalyzeStructure()
	if err == nil {
		t.Error("Expected error when root node doesn't exist")
	}

	_, err = invalidAnalyzer.GetNodeDistribution()
	if err == nil {
		t.Error("Expected error when root node doesn't exist for GetNodeDistribution")
	}

	_, err = invalidAnalyzer.CalculateHeight()
	if err == nil {
		t.Error("Expected error when root node doesn't exist for CalculateHeight")
	}

	_, err = invalidAnalyzer.CalculateFillFactor()
	if err == nil {
		t.Error("Expected error when root node doesn't exist for CalculateFillFactor")
	}
}

func TestBTreeAnalyzer_SingleNodeTree(t *testing.T) {
	// Create a single-node tree
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	rootData := createTestNavigatorNodeData(rootOID, types.BtnodeRoot|types.BtnodeLeaf, 0, 5, true)
	blockReader.SetBlock(types.Paddr(rootOID), rootData)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	analysis, err := analyzer.AnalyzeStructure()
	if err != nil {
		t.Fatalf("AnalyzeStructure on single node failed: %v", err)
	}

	if analysis.Height != 1 {
		t.Errorf("Expected height 1 for single node, got %d", analysis.Height)
	}

	if analysis.TotalNodes != 1 {
		t.Errorf("Expected 1 total node, got %d", analysis.TotalNodes)
	}

	if !analysis.IsBalanced {
		t.Error("Single node tree should be balanced")
	}

	if len(analysis.Levels) != 1 {
		t.Errorf("Expected 1 level, got %d", len(analysis.Levels))
	}
}

func BenchmarkBTreeAnalyzer_AnalyzeStructure(b *testing.B) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeStructure()
		if err != nil {
			b.Fatalf("AnalyzeStructure failed: %v", err)
		}
	}
}

func BenchmarkBTreeAnalyzer_GetNodeDistribution(b *testing.B) {
	navigator, btreeInfo, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)
	analyzer := NewBTreeAnalyzer(navigator, traverser, btreeInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.GetNodeDistribution()
		if err != nil {
			b.Fatalf("GetNodeDistribution failed: %v", err)
		}
	}
}
