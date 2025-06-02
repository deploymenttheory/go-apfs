package btrees

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
)

// btreeAnalyzer implements the BTreeAnalyzer interface
type btreeAnalyzer struct {
	navigator interfaces.BTreeNavigator
	traverser interfaces.BTreeTraverser
	btreeInfo interfaces.BTreeInfoReader
}

// NewBTreeAnalyzer creates a new BTreeAnalyzer implementation
func NewBTreeAnalyzer(navigator interfaces.BTreeNavigator, traverser interfaces.BTreeTraverser, btreeInfo interfaces.BTreeInfoReader) interfaces.BTreeAnalyzer {
	return &btreeAnalyzer{
		navigator: navigator,
		traverser: traverser,
		btreeInfo: btreeInfo,
	}
}

// GetNodeDistribution returns information about the distribution of nodes at each level
func (analyzer *btreeAnalyzer) GetNodeDistribution() ([]interfaces.LevelInfo, error) {
	levelMap := make(map[int][]interfaces.BTreeNodeReader)

	// Collect nodes by level using level-order traversal
	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		if levelMap[depth] == nil {
			levelMap[depth] = make([]interfaces.BTreeNodeReader, 0)
		}
		levelMap[depth] = append(levelMap[depth], node)
		return true, nil
	}

	err := analyzer.traverser.LevelOrderTraversal(visitor)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse levels: %w", err)
	}

	var levels []interfaces.LevelInfo

	// Process each level
	for level := 0; level < len(levelMap); level++ {
		nodes, exists := levelMap[level]
		if !exists || len(nodes) == 0 {
			break
		}

		totalKeys := 0
		minKeys := int(nodes[0].KeyCount())
		maxKeys := int(nodes[0].KeyCount())

		for _, node := range nodes {
			keyCount := int(node.KeyCount())
			totalKeys += keyCount

			if keyCount < minKeys {
				minKeys = keyCount
			}
			if keyCount > maxKeys {
				maxKeys = keyCount
			}
		}

		avgKeys := float64(totalKeys) / float64(len(nodes))

		levelInfo := interfaces.LevelInfo{
			Level:           level,
			NodeCount:       len(nodes),
			AverageKeyCount: avgKeys,
			MinKeyCount:     minKeys,
			MaxKeyCount:     maxKeys,
		}

		levels = append(levels, levelInfo)
	}

	return levels, nil
}

// CalculateFillFactor returns the average fill factor of the B-tree
func (analyzer *btreeAnalyzer) CalculateFillFactor() (float64, error) {
	nodeSize := analyzer.btreeInfo.NodeSize()
	if nodeSize == 0 {
		return 0, fmt.Errorf("invalid node size: %d", nodeSize)
	}

	var totalUsedSpace uint64
	var totalNodes int

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		totalNodes++

		// Calculate used space in this node
		nodeData := node.Data()
		usedSpace := uint64(len(nodeData))

		// Add header size (approximate)
		usedSpace += 56 // Size of btree_node_phys_t header

		totalUsedSpace += usedSpace
		return true, nil
	}

	err := analyzer.traverser.PreOrderTraversal(visitor)
	if err != nil {
		return 0, fmt.Errorf("failed to traverse tree: %w", err)
	}

	if totalNodes == 0 {
		return 0, nil
	}

	totalCapacity := uint64(totalNodes) * uint64(nodeSize)
	fillFactor := float64(totalUsedSpace) / float64(totalCapacity) * 100.0

	return fillFactor, nil
}

// CalculateHeight returns the height of the B-tree
func (analyzer *btreeAnalyzer) CalculateHeight() (int, error) {
	height, err := analyzer.navigator.GetHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to get tree height: %w", err)
	}

	return int(height), nil
}

// AnalyzeStructure performs a comprehensive analysis of the B-tree structure
func (analyzer *btreeAnalyzer) AnalyzeStructure() (interfaces.BTreeAnalysis, error) {
	// Get height
	height, err := analyzer.CalculateHeight()
	if err != nil {
		return interfaces.BTreeAnalysis{}, fmt.Errorf("failed to calculate height: %w", err)
	}

	// Get node distribution
	levels, err := analyzer.GetNodeDistribution()
	if err != nil {
		return interfaces.BTreeAnalysis{}, fmt.Errorf("failed to get node distribution: %w", err)
	}

	// Calculate totals
	totalNodes := 0
	totalKeys := 0
	for _, level := range levels {
		totalNodes += level.NodeCount
		totalKeys += int(float64(level.NodeCount) * level.AverageKeyCount)
	}

	// Get fill factor
	fillFactor, err := analyzer.CalculateFillFactor()
	if err != nil {
		return interfaces.BTreeAnalysis{}, fmt.Errorf("failed to calculate fill factor: %w", err)
	}

	// Analyze balance
	isBalanced := analyzer.analyzeBalance(levels)

	// Find largest and smallest nodes
	largestNode, smallestNode, err := analyzer.findExtremeNodes()
	if err != nil {
		return interfaces.BTreeAnalysis{}, fmt.Errorf("failed to find extreme nodes: %w", err)
	}

	return interfaces.BTreeAnalysis{
		Height:       height,
		TotalNodes:   totalNodes,
		TotalKeys:    totalKeys,
		FillFactor:   fillFactor,
		Levels:       levels,
		IsBalanced:   isBalanced,
		LargestNode:  largestNode,
		SmallestNode: smallestNode,
	}, nil
}

// analyzeBalance determines if the tree is balanced
func (analyzer *btreeAnalyzer) analyzeBalance(levels []interfaces.LevelInfo) bool {
	if len(levels) <= 1 {
		return true // Single level is always balanced
	}

	// Check if all leaf nodes are at the same level
	// In a balanced B-tree, all leaves should be at the deepest level
	leafLevel := len(levels) - 1

	// Simple heuristic: if the last level has nodes, it's likely balanced
	// A more sophisticated check would verify that all leaves are at the same depth
	return levels[leafLevel].NodeCount > 0
}

// findExtremeNodes finds the largest and smallest nodes by key count
func (analyzer *btreeAnalyzer) findExtremeNodes() (interfaces.BTreeNodeInfo, interfaces.BTreeNodeInfo, error) {
	var largestNode, smallestNode interfaces.BTreeNodeInfo
	found := false

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		nodeInfo := interfaces.BTreeNodeInfo{
			ObjectID:    0, // We don't have access to OID in the interface
			Level:       node.Level(),
			KeyCount:    node.KeyCount(),
			Flags:       node.Flags(),
			SizeInBytes: uint32(len(node.Data()) + 56), // Data + header size
		}

		if !found {
			largestNode = nodeInfo
			smallestNode = nodeInfo
			found = true
		} else {
			if nodeInfo.KeyCount > largestNode.KeyCount {
				largestNode = nodeInfo
			}
			if nodeInfo.KeyCount < smallestNode.KeyCount {
				smallestNode = nodeInfo
			}
		}

		return true, nil
	}

	err := analyzer.traverser.PreOrderTraversal(visitor)
	if err != nil {
		return interfaces.BTreeNodeInfo{}, interfaces.BTreeNodeInfo{}, fmt.Errorf("failed to traverse tree: %w", err)
	}

	if !found {
		return interfaces.BTreeNodeInfo{}, interfaces.BTreeNodeInfo{}, fmt.Errorf("no nodes found in tree")
	}

	return largestNode, smallestNode, nil
}

// AnalyzeKeyDistribution analyzes the distribution of keys across the tree
func (analyzer *btreeAnalyzer) AnalyzeKeyDistribution() (KeyDistributionAnalysis, error) {
	var totalKeys uint64
	var leafKeys uint64
	var internalKeys uint64
	var minKeysPerNode uint32 = ^uint32(0) // Max uint32
	var maxKeysPerNode uint32
	nodeCount := 0
	leafCount := 0

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		keyCount := node.KeyCount()
		totalKeys += uint64(keyCount)
		nodeCount++

		if keyCount < minKeysPerNode {
			minKeysPerNode = keyCount
		}
		if keyCount > maxKeysPerNode {
			maxKeysPerNode = keyCount
		}

		if node.IsLeaf() {
			leafKeys += uint64(keyCount)
			leafCount++
		} else {
			internalKeys += uint64(keyCount)
		}

		return true, nil
	}

	err := analyzer.traverser.PreOrderTraversal(visitor)
	if err != nil {
		return KeyDistributionAnalysis{}, fmt.Errorf("failed to traverse tree: %w", err)
	}

	if nodeCount == 0 {
		minKeysPerNode = 0
	}

	avgKeysPerNode := float64(totalKeys) / float64(nodeCount)
	avgKeysPerLeaf := float64(leafKeys) / float64(leafCount)

	return KeyDistributionAnalysis{
		TotalKeys:         totalKeys,
		LeafKeys:          leafKeys,
		InternalKeys:      internalKeys,
		MinKeysPerNode:    minKeysPerNode,
		MaxKeysPerNode:    maxKeysPerNode,
		AvgKeysPerNode:    avgKeysPerNode,
		AvgKeysPerLeaf:    avgKeysPerLeaf,
		LeafNodeCount:     leafCount,
		InternalNodeCount: nodeCount - leafCount,
	}, nil
}

// AnalyzeStorageEfficiency analyzes how efficiently the tree uses storage
func (analyzer *btreeAnalyzer) AnalyzeStorageEfficiency() (StorageEfficiencyAnalysis, error) {
	nodeSize := analyzer.btreeInfo.NodeSize()
	keySize := analyzer.btreeInfo.KeySize()
	valueSize := analyzer.btreeInfo.ValueSize()

	var wastedSpace uint64
	var usedSpace uint64
	nodeCount := 0

	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		nodeCount++

		// Calculate theoretical space usage
		keyCount := node.KeyCount()
		if keySize > 0 && valueSize > 0 {
			// Fixed-size keys and values
			theoreticalDataSize := uint64(keyCount) * uint64(keySize+valueSize)
			actualDataSize := uint64(len(node.Data()))

			if actualDataSize > theoreticalDataSize {
				wastedSpace += actualDataSize - theoreticalDataSize
			}
			usedSpace += actualDataSize
		} else {
			// Variable-size keys/values - just count actual usage
			usedSpace += uint64(len(node.Data()))
		}

		return true, nil
	}

	err := analyzer.traverser.PreOrderTraversal(visitor)
	if err != nil {
		return StorageEfficiencyAnalysis{}, fmt.Errorf("failed to traverse tree: %w", err)
	}

	totalCapacity := uint64(nodeCount) * uint64(nodeSize)
	utilization := float64(usedSpace) / float64(totalCapacity) * 100.0
	wastePercentage := float64(wastedSpace) / float64(usedSpace) * 100.0

	return StorageEfficiencyAnalysis{
		TotalCapacity:   totalCapacity,
		UsedSpace:       usedSpace,
		WastedSpace:     wastedSpace,
		Utilization:     utilization,
		WastePercentage: wastePercentage,
		NodeCount:       nodeCount,
		AvgNodeSize:     float64(usedSpace) / float64(nodeCount),
	}, nil
}

// KeyDistributionAnalysis contains analysis of key distribution
type KeyDistributionAnalysis struct {
	TotalKeys         uint64
	LeafKeys          uint64
	InternalKeys      uint64
	MinKeysPerNode    uint32
	MaxKeysPerNode    uint32
	AvgKeysPerNode    float64
	AvgKeysPerLeaf    float64
	LeafNodeCount     int
	InternalNodeCount int
}

// StorageEfficiencyAnalysis contains analysis of storage efficiency
type StorageEfficiencyAnalysis struct {
	TotalCapacity   uint64
	UsedSpace       uint64
	WastedSpace     uint64
	Utilization     float64
	WastePercentage float64
	NodeCount       int
	AvgNodeSize     float64
}
