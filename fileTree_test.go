package dfx

import (
	"testing"
)

// helper to build a test tree:
//
//	root/
//	  src/
//	    main.go
//	    util.go
//	    README.md
//	  docs/
//	    guide.txt
//	  Makefile
func testTree() *FileNode {
	root := &FileNode{Name: "root", Dir: true}

	src := &FileNode{Name: "src", Dir: true, Parent: root}
	mainGo := &FileNode{Name: "main.go", Parent: src}
	utilGo := &FileNode{Name: "util.go", Parent: src}
	readme := &FileNode{Name: "README.md", Parent: src}
	src.Children = []*FileNode{mainGo, utilGo, readme}

	docs := &FileNode{Name: "docs", Dir: true, Parent: root}
	guide := &FileNode{Name: "guide.txt", Parent: docs}
	docs.Children = []*FileNode{guide}

	makefile := &FileNode{Name: "Makefile", Parent: root}

	root.Children = []*FileNode{src, docs, makefile}
	return root
}

func TestFind_NilReceiver(t *testing.T) {
	var n *FileNode
	results := n.Find(func(*FileNode) bool { return true })
	if results != nil {
		t.Fatalf("expected nil, got %v", results)
	}
}

func TestFind_SingleNodeMatch(t *testing.T) {
	node := &FileNode{Name: "test.go"}
	results := node.Find(func(n *FileNode) bool { return n.Name == "test.go" })
	if len(results) != 1 || results[0] != node {
		t.Fatalf("expected single match, got %d", len(results))
	}
}

func TestFind_SingleNodeNoMatch(t *testing.T) {
	node := &FileNode{Name: "test.go"}
	results := node.Find(func(n *FileNode) bool { return n.Name == "other.go" })
	if len(results) != 0 {
		t.Fatalf("expected no matches, got %d", len(results))
	}
}

func TestFind_DepthFirstOrder(t *testing.T) {
	root := testTree()
	// collect all node names in traversal order
	var names []string
	root.Find(func(n *FileNode) bool {
		names = append(names, n.Name)
		return false // don't care about results, just collecting order
	})

	expected := []string{"root", "src", "main.go", "util.go", "README.md", "docs", "guide.txt", "Makefile"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d nodes, got %d: %v", len(expected), len(names), names)
	}
	for i, name := range names {
		if name != expected[i] {
			t.Fatalf("position %d: expected '%s', got '%s'", i, expected[i], name)
		}
	}
}

func TestFind_DirectoriesOnly(t *testing.T) {
	root := testTree()
	dirs := root.Find(func(n *FileNode) bool { return n.Dir })
	if len(dirs) != 3 {
		t.Fatalf("expected 3 directories, got %d", len(dirs))
	}
	expected := []string{"root", "src", "docs"}
	for i, d := range dirs {
		if d.Name != expected[i] {
			t.Fatalf("position %d: expected '%s', got '%s'", i, expected[i], d.Name)
		}
	}
}

func TestFind_FilesOnly(t *testing.T) {
	root := testTree()
	files := root.Find(func(n *FileNode) bool { return !n.Dir })
	if len(files) != 5 {
		t.Fatalf("expected 5 files, got %d", len(files))
	}
	expected := []string{"main.go", "util.go", "README.md", "guide.txt", "Makefile"}
	for i, f := range files {
		if f.Name != expected[i] {
			t.Fatalf("position %d: expected '%s', got '%s'", i, expected[i], f.Name)
		}
	}
}

func TestFind_NoMatches(t *testing.T) {
	root := testTree()
	results := root.Find(func(n *FileNode) bool { return n.Name == "nonexistent" })
	if len(results) != 0 {
		t.Fatalf("expected no matches, got %d", len(results))
	}
}

func TestMatchExt(t *testing.T) {
	root := testTree()
	goFiles := root.Find(MatchExt(".go"))
	if len(goFiles) != 2 {
		t.Fatalf("expected 2 .go files, got %d", len(goFiles))
	}
	if goFiles[0].Name != "main.go" || goFiles[1].Name != "util.go" {
		t.Fatalf("unexpected files: %s, %s", goFiles[0].Name, goFiles[1].Name)
	}
}

func TestMatchExt_ExcludesDirectories(t *testing.T) {
	// create a directory that ends with .go
	root := &FileNode{Name: "root", Dir: true}
	goDir := &FileNode{Name: "pkg.go", Dir: true, Parent: root}
	goFile := &FileNode{Name: "main.go", Parent: root}
	root.Children = []*FileNode{goDir, goFile}

	results := root.Find(MatchExt(".go"))
	if len(results) != 1 {
		t.Fatalf("expected 1 match (file only), got %d", len(results))
	}
	if results[0].Name != "main.go" {
		t.Fatalf("expected 'main.go', got '%s'", results[0].Name)
	}
}

func TestMatchExt_CaseInsensitive(t *testing.T) {
	root := &FileNode{Name: "root", Dir: true}
	upper := &FileNode{Name: "README.GO", Parent: root}
	mixed := &FileNode{Name: "test.Go", Parent: root}
	root.Children = []*FileNode{upper, mixed}

	results := root.Find(MatchExt(".go"))
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestMatchName_ValidPattern(t *testing.T) {
	root := testTree()
	pred, err := MatchName(`^main\.`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := root.Find(pred)
	if len(results) != 1 || results[0].Name != "main.go" {
		t.Fatalf("expected 1 match for 'main.go', got %d", len(results))
	}
}

func TestMatchName_MultipleMatches(t *testing.T) {
	root := testTree()
	pred, err := MatchName(`\.go$`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := root.Find(pred)
	if len(results) != 2 {
		t.Fatalf("expected 2 .go matches, got %d", len(results))
	}
}

func TestMatchName_InvalidRegex(t *testing.T) {
	_, err := MatchName(`[invalid`)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestMatchPath_MatchesFullPath(t *testing.T) {
	root := testTree()
	pred, err := MatchPath(`src/main\.go`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := root.Find(pred)
	if len(results) != 1 || results[0].Name != "main.go" {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestMatchPath_MatchesDirectory(t *testing.T) {
	root := testTree()
	pred, err := MatchPath(`^docs$`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := root.Find(pred)
	if len(results) != 1 || results[0].Name != "docs" {
		t.Fatalf("expected 1 match for 'docs', got %d", len(results))
	}
}

func TestMatchPath_InvalidRegex(t *testing.T) {
	_, err := MatchPath(`[invalid`)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}
