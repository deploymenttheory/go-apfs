package btrees

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestNewBTreeTraverser(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	if traverser == nil {
		t.Fatal("NewBTreeTraverser returned nil")
	}
}

func TestBTreeTraverser_PreOrderTraversal(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitedNodes := 0
	err := traverser.PreOrderTraversal(func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		visitedNodes++
		return true, nil // Continue traversal
	})

	if err != nil {
		t.Fatalf("PreOrderTraversal failed: %v", err)
	}

	// Should visit root + 4 leaf nodes = 5 nodes (changed from 4 to 5)
	if visitedNodes != 5 {
		t.Errorf("Visited %d nodes, expected 5", visitedNodes)
	}
}

func TestBTreeTraverser_InOrderTraversal(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitedNodes := []uint32{}

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		visitedNodes = append(visitedNodes, node.KeyCount())
		return true, nil
	}

	err := traverser.InOrderTraversal(visitor)
	if err != nil {
		t.Fatalf("InOrderTraversal failed: %v", err)
	}

	// In-order traversal visits nodes in a different order than pre-order
	if len(visitedNodes) == 0 {
		t.Error("No nodes visited during in-order traversal")
	}
}

func TestBTreeTraverser_PostOrderTraversal(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitedNodes := []uint32{}
	visitedDepths := []int{}

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		visitedNodes = append(visitedNodes, node.KeyCount())
		visitedDepths = append(visitedDepths, depth)
		return true, nil
	}

	err := traverser.PostOrderTraversal(visitor)
	if err != nil {
		t.Fatalf("PostOrderTraversal failed: %v", err)
	}

	// Expected: 4 leaf nodes (3 keys each), then root (3 keys) = 5 total nodes
	expectedNodes := []uint32{3, 3, 3, 3, 3} // Changed from 4 to 5 nodes

	if len(visitedNodes) != len(expectedNodes) {
		t.Errorf("Visited %d nodes, expected %d", len(visitedNodes), len(expectedNodes))
	}

	// Check that the last visited node is the root (depth 0)
	if len(visitedDepths) > 0 && visitedDepths[len(visitedDepths)-1] != 0 {
		t.Error("Root should be visited last in post-order traversal")
	}
}

func TestBTreeTraverser_LevelOrderTraversal(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitedNodes := []uint32{}
	visitedDepths := []int{}

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		visitedNodes = append(visitedNodes, node.KeyCount())
		visitedDepths = append(visitedDepths, depth)
		return true, nil
	}

	err := traverser.LevelOrderTraversal(visitor)
	if err != nil {
		t.Fatalf("LevelOrderTraversal failed: %v", err)
	}

	// Level-order should visit root first, then all leaves = 5 total nodes
	expectedNodes := []uint32{3, 3, 3, 3, 3} // Changed from 4 to 5 nodes

	if len(visitedNodes) != len(expectedNodes) {
		t.Errorf("Visited %d nodes, expected %d", len(visitedNodes), len(expectedNodes))
	}

	// Check depth progression (should be non-decreasing)
	for i := 1; i < len(visitedDepths); i++ {
		if visitedDepths[i] < visitedDepths[i-1] {
			t.Errorf("Depth decreased from %d to %d at position %d", visitedDepths[i-1], visitedDepths[i], i)
		}
	}
}

func TestBTreeTraverser_VisitorEarlyExit(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitCount := 0
	maxVisits := 1 // Changed to 1 so it stops after visiting the root

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		visitCount++
		// Stop after visiting maxVisits nodes
		return visitCount < maxVisits, nil
	}

	err := traverser.PreOrderTraversal(visitor)
	if err != nil {
		t.Fatalf("PreOrderTraversal with early exit failed: %v", err)
	}

	// With maxVisits = 1, it should visit only the root node and then stop
	// because the visitor returns false, preventing traversal of children
	if visitCount != 1 {
		t.Errorf("Expected to visit 1 node, but visited %d", visitCount)
	}
}

func TestBTreeTraverser_AdditionalMethods(t *testing.T) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator).(*btreeTraverser)

	// Test TraverseLeaves
	leafCount := 0
	leafVisitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		if !node.IsLeaf() {
			t.Error("TraverseLeaves visited a non-leaf node")
		}
		leafCount++
		return true, nil
	}

	err := traverser.TraverseLeaves(leafVisitor)
	if err != nil {
		t.Fatalf("TraverseLeaves failed: %v", err)
	}

	if leafCount != 4 { // Changed from 3 to 4 leaf nodes
		t.Errorf("Expected to visit 4 leaf nodes, visited %d", leafCount)
	}

	// Test GetNodeCount
	nodeCount, err := traverser.GetNodeCount()
	if err != nil {
		t.Fatalf("GetNodeCount failed: %v", err)
	}

	if nodeCount != 5 { // 1 root + 4 leaves (changed from 4 to 5)
		t.Errorf("Expected 5 nodes, got %d", nodeCount)
	}

	// Test GetLeafCount
	leafCount2, err := traverser.GetLeafCount()
	if err != nil {
		t.Fatalf("GetLeafCount failed: %v", err)
	}

	if leafCount2 != 4 { // Changed from 3 to 4 leaf nodes
		t.Errorf("Expected 4 leaf nodes, got %d", leafCount2)
	}

	// Test GetMaxDepth
	maxDepth, err := traverser.GetMaxDepth()
	if err != nil {
		t.Fatalf("GetMaxDepth failed: %v", err)
	}

	if maxDepth != 1 { // Leaves are at depth 1
		t.Errorf("Expected max depth 1, got %d", maxDepth)
	}
}

func TestBTreeTraverser_ErrorCases(t *testing.T) {
	// Test with corrupted tree (missing root)
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(9999) // Non-existent root

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)
	traverser := NewBTreeTraverser(navigator)

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		return true, nil
	}

	err := traverser.PreOrderTraversal(visitor)
	if err == nil {
		t.Error("Expected error when root node doesn't exist")
	}

	err = traverser.InOrderTraversal(visitor)
	if err == nil {
		t.Error("Expected error when root node doesn't exist for InOrderTraversal")
	}

	err = traverser.PostOrderTraversal(visitor)
	if err == nil {
		t.Error("Expected error when root node doesn't exist for PostOrderTraversal")
	}

	err = traverser.LevelOrderTraversal(visitor)
	if err == nil {
		t.Error("Expected error when root node doesn't exist for LevelOrderTraversal")
	}
}

func TestBTreeTraverser_SingleNodeTree(t *testing.T) {
	// Create a tree with just a root node (leaf)
	blockReader := NewMockBlockDeviceReader()
	btreeInfo := NewMockBTreeInfoReader()
	rootOID := types.OidT(1000)

	rootData := createTestNavigatorNodeData(rootOID, types.BtnodeRoot|types.BtnodeLeaf, 0, 5, true)
	blockReader.SetBlock(types.Paddr(rootOID), rootData)

	navigator := NewBTreeNavigator(blockReader, rootOID, btreeInfo)
	traverser := NewBTreeTraverser(navigator)

	visitCount := 0
	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		visitCount++
		if depth != 0 {
			t.Errorf("Expected depth 0 for single node tree, got %d", depth)
		}
		if !node.IsRoot() || !node.IsLeaf() {
			t.Error("Single node should be both root and leaf")
		}
		return true, nil
	}

	// Test all traversal methods on single node
	err := traverser.PreOrderTraversal(visitor)
	if err != nil {
		t.Fatalf("PreOrderTraversal on single node failed: %v", err)
	}
	if visitCount != 1 {
		t.Errorf("Expected 1 visit for single node, got %d", visitCount)
	}

	visitCount = 0
	err = traverser.LevelOrderTraversal(visitor)
	if err != nil {
		t.Fatalf("LevelOrderTraversal on single node failed: %v", err)
	}
	if visitCount != 1 {
		t.Errorf("Expected 1 visit for single node, got %d", visitCount)
	}
}

func BenchmarkBTreeTraverser_PreOrderTraversal(b *testing.B) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		// Simple visitor that just continues
		return true, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := traverser.PreOrderTraversal(visitor)
		if err != nil {
			b.Fatalf("PreOrderTraversal failed: %v", err)
		}
	}
}

func BenchmarkBTreeTraverser_LevelOrderTraversal(b *testing.B) {
	navigator, _, _ := createTestSearchTree()
	traverser := NewBTreeTraverser(navigator)

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		return true, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := traverser.LevelOrderTraversal(visitor)
		if err != nil {
			b.Fatalf("LevelOrderTraversal failed: %v", err)
		}
	}
}
