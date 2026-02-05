package dfx

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
)

// FileNode represents a node in the filesystem tree.
// nodes maintain parent-child relationships and can represent
// either directories or files.
type FileNode struct {
	Name     string
	Dir      bool
	Parent   *FileNode
	Children []*FileNode
}

// Path returns the full filesystem path from the root to this node.
func (n *FileNode) Path() string {
	if n == nil {
		return ""
	}
	if n.Parent != nil {
		return filepath.Join(n.Parent.Path(), n.Name)
	}
	return n.Name
}

// BuildTree recursively scans a filesystem path and builds a tree structure.
// the parent parameter should be nil for the root node.
func BuildTree(path string, parent *FileNode) (*FileNode, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot stat '%v': %w", path, err)
	}

	node := &FileNode{
		Name:   fi.Name(),
		Dir:    fi.IsDir(),
		Parent: parent,
	}

	if node.Dir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read directory '%v': %w", path, err)
		}

		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			child, err := BuildTree(childPath, node)
			if err != nil {
				return nil, fmt.Errorf("cannot build tree for '%v': %w", childPath, err)
			}
			node.Children = append(node.Children, child)
		}
	}

	return node, nil
}

// Find returns all nodes in the tree rooted at this node for which
// the predicate returns true. traversal is depth-first pre-order.
func (n *FileNode) Find(predicate func(*FileNode) bool) []*FileNode {
	if n == nil {
		return nil
	}
	var results []*FileNode
	n.findRecursive(predicate, &results)
	return results
}

func (n *FileNode) findRecursive(predicate func(*FileNode) bool, results *[]*FileNode) {
	if predicate(n) {
		*results = append(*results, n)
	}
	for _, child := range n.Children {
		child.findRecursive(predicate, results)
	}
}

// MatchExt returns a predicate that matches non-directory nodes whose
// file extension equals ext. the ext parameter should include the dot
// (e.g. ".go"). the comparison is case-insensitive.
func MatchExt(ext string) func(*FileNode) bool {
	ext = strings.ToLower(ext)
	return func(n *FileNode) bool {
		if n.Dir {
			return false
		}
		return strings.ToLower(filepath.Ext(n.Name)) == ext
	}
}

// MatchName returns a predicate that matches nodes whose Name matches
// the given regular expression pattern.
func MatchName(pattern string) (func(*FileNode) bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid name pattern '%v': %w", pattern, err)
	}
	return func(n *FileNode) bool {
		return re.MatchString(n.Name)
	}, nil
}

// MatchPath returns a predicate that matches nodes whose full Path()
// matches the given regular expression pattern.
func MatchPath(pattern string) (func(*FileNode) bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid path pattern '%v': %w", pattern, err)
	}
	return func(n *FileNode) bool {
		return re.MatchString(n.Path())
	}, nil
}

// FileTree is a component that displays a filesystem tree with
// selection and interaction support.
type FileTree struct {
	Container
	Root          *FileNode
	Selected      *FileNode
	OnSelect      func(*FileNode)
	OnDoubleClick func(*FileNode)
	Filter        func(*FileNode) bool
}

// NewFileTree creates a new filesystem tree component.
func NewFileTree(root *FileNode) *FileTree {
	return &FileTree{
		Container: Container{Visible: true},
		Root:      root,
	}
}

// SetRoot updates the tree root and clears selection.
func (ft *FileTree) SetRoot(root *FileNode) {
	ft.Root = root
	ft.Selected = nil
}

// SelectNode programmatically selects a node.
func (ft *FileTree) SelectNode(node *FileNode) {
	ft.Selected = node
	if ft.OnSelect != nil {
		ft.OnSelect(node)
	}
}

// Draw renders the filesystem tree.
func (ft *FileTree) Draw(state *State) {
	if !ft.Visible {
		return
	}

	// create scrollable child window for the tree
	imgui.BeginChildStrV("##fileTree", imgui.Vec2{X: 0, Y: 0}, 0, imgui.WindowFlagsHorizontalScrollbar)

	// render the tree recursively
	ft.visitNode(ft.Root)

	imgui.EndChild()

	// call base container drawing
	if ft.OnDraw != nil {
		ft.OnDraw(state)
	}
	for _, child := range ft.Children {
		child.Draw(state)
	}
}

// visitNode recursively renders a single node and its children.
func (ft *FileTree) visitNode(node *FileNode) {
	if node == nil {
		return
	}

	// check filter if provided
	if ft.Filter != nil && !ft.Filter(node) {
		return
	}

	// base flags for all tree nodes
	baseFlags := imgui.TreeNodeFlagsOpenOnArrow |
		imgui.TreeNodeFlagsOpenOnDoubleClick |
		imgui.TreeNodeFlagsSpanAvailWidth

	if node.Dir {
		// render directory node
		if imgui.TreeNodeExStrV(node.Name, baseFlags) {
			// render children recursively
			for _, child := range node.Children {
				ft.visitNode(child)
			}
			imgui.TreePop()
		}

		// handle directory selection
		if imgui.IsItemClicked() {
			ft.Selected = node
			if ft.OnSelect != nil {
				ft.OnSelect(node)
			}
		}

		// handle directory double-click
		if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
			if ft.OnDoubleClick != nil {
				ft.OnDoubleClick(node)
			}
		}
	} else {
		// render file node (leaf)
		leafFlags := baseFlags | imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsNoTreePushOnOpen

		// highlight if selected
		if node == ft.Selected {
			leafFlags |= imgui.TreeNodeFlagsSelected
		}

		imgui.TreeNodeExStrV(node.Name, leafFlags)

		// handle file selection
		if imgui.IsItemClicked() {
			ft.Selected = node
			if ft.OnSelect != nil {
				ft.OnSelect(node)
			}
		}

		// handle file double-click
		if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) {
			if ft.OnDoubleClick != nil {
				ft.OnDoubleClick(node)
			}
		}
	}
}
