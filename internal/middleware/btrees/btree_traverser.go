package btrees

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
)

// btreeTraverser implements the BTreeTraverser interface
type btreeTraverser struct {
	navigator interfaces.BTreeNavigator
}

// NewBTreeTraverser creates a new BTreeTraverser implementation
func NewBTreeTraverser(navigator interfaces.BTreeNavigator) interfaces.BTreeTraverser {
	return &btreeTraverser{
		navigator: navigator,
	}
}

// PreOrderTraversal performs a pre-order traversal of the B-tree
func (traverser *btreeTraverser) PreOrderTraversal(visitor interfaces.NodeVisitor) error {
	rootNode, err := traverser.navigator.GetRootNode()
	if err != nil {
		return fmt.Errorf("failed to get root node: %w", err)
	}

	return traverser.preOrderTraversal(rootNode, 0, visitor)
}

// InOrderTraversal performs an in-order traversal of the B-tree
func (traverser *btreeTraverser) InOrderTraversal(visitor interfaces.NodeVisitor) error {
	rootNode, err := traverser.navigator.GetRootNode()
	if err != nil {
		return fmt.Errorf("failed to get root node: %w", err)
	}

	return traverser.inOrderTraversal(rootNode, 0, visitor)
}

// PostOrderTraversal performs a post-order traversal of the B-tree
func (traverser *btreeTraverser) PostOrderTraversal(visitor interfaces.NodeVisitor) error {
	rootNode, err := traverser.navigator.GetRootNode()
	if err != nil {
		return fmt.Errorf("failed to get root node: %w", err)
	}

	return traverser.postOrderTraversal(rootNode, 0, visitor)
}

// LevelOrderTraversal performs a level-order traversal of the B-tree
func (traverser *btreeTraverser) LevelOrderTraversal(visitor interfaces.NodeVisitor) error {
	rootNode, err := traverser.navigator.GetRootNode()
	if err != nil {
		return fmt.Errorf("failed to get root node: %w", err)
	}

	return traverser.levelOrderTraversal(rootNode, visitor)
}

// preOrderTraversal recursively performs pre-order traversal
func (traverser *btreeTraverser) preOrderTraversal(node interfaces.BTreeNodeReader, depth int, visitor interfaces.NodeVisitor) error {
	// Visit the current node first
	shouldContinue, err := visitor(node, depth)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil // Stop traversal
	}

	// Then visit children (if not a leaf)
	if !node.IsLeaf() {
		keyCount := int(node.KeyCount())
		// In B-trees, internal nodes have keyCount + 1 children
		childCount := keyCount + 1

		for i := 0; i < childCount; i++ {
			childNode, err := traverser.navigator.GetChildNode(node, i)
			if err != nil {
				return fmt.Errorf("failed to get child node %d: %w", i, err)
			}

			if err := traverser.preOrderTraversal(childNode, depth+1, visitor); err != nil {
				return err
			}
		}
	}

	return nil
}

// inOrderTraversal recursively performs in-order traversal
func (traverser *btreeTraverser) inOrderTraversal(node interfaces.BTreeNodeReader, depth int, visitor interfaces.NodeVisitor) error {
	if node.IsLeaf() {
		// For leaf nodes, just visit the node
		_, err := visitor(node, depth)
		return err
	}

	// For internal nodes, interleave children and keys
	keyCount := int(node.KeyCount())

	for i := 0; i <= keyCount; i++ {
		// Visit left child
		if i < keyCount {
			childNode, err := traverser.navigator.GetChildNode(node, i)
			if err != nil {
				return fmt.Errorf("failed to get child node %d: %w", i, err)
			}

			if err := traverser.inOrderTraversal(childNode, depth+1, visitor); err != nil {
				return err
			}
		}

		// Visit current node (representing the key at position i)
		if i == keyCount/2 { // Visit the node at the middle position
			shouldContinue, err := visitor(node, depth)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
		}
	}

	return nil
}

// postOrderTraversal recursively performs post-order traversal
func (traverser *btreeTraverser) postOrderTraversal(node interfaces.BTreeNodeReader, depth int, visitor interfaces.NodeVisitor) error {
	// Visit children first (if not a leaf)
	if !node.IsLeaf() {
		keyCount := int(node.KeyCount())
		childCount := keyCount + 1

		for i := 0; i < childCount; i++ {
			childNode, err := traverser.navigator.GetChildNode(node, i)
			if err != nil {
				return fmt.Errorf("failed to get child node %d: %w", i, err)
			}

			if err := traverser.postOrderTraversal(childNode, depth+1, visitor); err != nil {
				return err
			}
		}
	}

	// Then visit the current node
	_, err := visitor(node, depth)
	return err
}

// levelOrderTraversal performs level-order (breadth-first) traversal
func (traverser *btreeTraverser) levelOrderTraversal(rootNode interfaces.BTreeNodeReader, visitor interfaces.NodeVisitor) error {
	// Use a queue to track nodes to visit
	type nodeDepthPair struct {
		node  interfaces.BTreeNodeReader
		depth int
	}

	queue := []nodeDepthPair{{node: rootNode, depth: 0}}

	for len(queue) > 0 {
		// Dequeue the front node
		current := queue[0]
		queue = queue[1:]

		// Visit the current node
		shouldContinue, err := visitor(current.node, current.depth)
		if err != nil {
			return err
		}
		if !shouldContinue {
			continue // Skip adding children to queue
		}

		// Add children to queue (if not a leaf)
		if !current.node.IsLeaf() {
			keyCount := int(current.node.KeyCount())
			childCount := keyCount + 1

			for i := 0; i < childCount; i++ {
				childNode, err := traverser.navigator.GetChildNode(current.node, i)
				if err != nil {
					return fmt.Errorf("failed to get child node %d: %w", i, err)
				}

				queue = append(queue, nodeDepthPair{
					node:  childNode,
					depth: current.depth + 1,
				})
			}
		}
	}

	return nil
}

// TraverseLeaves traverses only the leaf nodes of the B-tree
func (traverser *btreeTraverser) TraverseLeaves(visitor interfaces.NodeVisitor) error {
	leafVisitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		if node.IsLeaf() {
			return visitor(node, depth)
		}
		return true, nil // Continue traversal for internal nodes
	}

	return traverser.PreOrderTraversal(leafVisitor)
}

// TraverseByLevel traverses nodes level by level, calling the visitor for each level
func (traverser *btreeTraverser) TraverseByLevel(levelVisitor func(level int, nodes []interfaces.BTreeNodeReader) error) error {
	rootNode, err := traverser.navigator.GetRootNode()
	if err != nil {
		return fmt.Errorf("failed to get root node: %w", err)
	}

	currentLevel := []interfaces.BTreeNodeReader{rootNode}
	level := 0

	for len(currentLevel) > 0 {
		// Visit all nodes at the current level
		if err := levelVisitor(level, currentLevel); err != nil {
			return err
		}

		// Prepare next level
		var nextLevel []interfaces.BTreeNodeReader
		for _, node := range currentLevel {
			if !node.IsLeaf() {
				keyCount := int(node.KeyCount())
				childCount := keyCount + 1

				for i := 0; i < childCount; i++ {
					childNode, err := traverser.navigator.GetChildNode(node, i)
					if err != nil {
						return fmt.Errorf("failed to get child node %d at level %d: %w", i, level, err)
					}
					nextLevel = append(nextLevel, childNode)
				}
			}
		}

		currentLevel = nextLevel
		level++
	}

	return nil
}

// GetNodeCount returns the total number of nodes in the B-tree
func (traverser *btreeTraverser) GetNodeCount() (int, error) {
	count := 0
	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		count++
		return true, nil
	}

	err := traverser.PreOrderTraversal(visitor)
	return count, err
}

// GetLeafCount returns the number of leaf nodes in the B-tree
func (traverser *btreeTraverser) GetLeafCount() (int, error) {
	count := 0
	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		if node.IsLeaf() {
			count++
		}
		return true, nil
	}

	err := traverser.PreOrderTraversal(visitor)
	return count, err
}

// GetMaxDepth returns the maximum depth of the B-tree
func (traverser *btreeTraverser) GetMaxDepth() (int, error) {
	maxDepth := 0
	visitor := func(node interfaces.BTreeNodeReader, depth int) (bool, error) {
		if depth > maxDepth {
			maxDepth = depth
		}
		return true, nil
	}

	err := traverser.PreOrderTraversal(visitor)
	return maxDepth, err
}
