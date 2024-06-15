package loader

import (
	"path/filepath"
)

// FilePath helps w/ type safety and readability
type FilePath string

type MyTreeNode interface {
	Parent() MyTreeNode
	Name() string
	Path() FilePath
	Root() MyTreeNode
	Accept(TreeVisitor)
}

// myTreeNode is the commonality between a file and a folder
type myTreeNode struct {
	parent MyTreeNode
	name   string
}

var _ MyTreeNode = &myTreeNode{}

// Root returns the "highest" non-nil tree item.
func (ti *myTreeNode) Root() MyTreeNode {
	if ti == nil {
		return nil
	}
	if ti.parent == nil {
		// This is how it stops.
		return ti
	}
	return ti.parent.Root()
}

// Name is the base name of the item.
func (ti *myTreeNode) Name() string {
	if ti == nil {
		return ""
	}
	return ti.name
}

// Path is the fully qualified name of the item, including parents.
func (ti *myTreeNode) Path() FilePath {
	if ti == nil {
		return RootSlash
	}
	if ti.parent == nil {
		return FilePath(ti.name)
	}
	return FilePath(filepath.Join(string(ti.parent.Path()), ti.name))
}

// Parent is the parent of the item.
func (ti *myTreeNode) Parent() MyTreeNode {
	return ti.parent
}

func (ti *myTreeNode) Accept(_ TreeVisitor) {
	// overridden in subclasses
}
