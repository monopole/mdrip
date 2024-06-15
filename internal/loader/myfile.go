package loader

// MyFile is named byte array.
type MyFile struct {
	myTreeNode
	content []byte
}

var _ MyTreeNode = &MyFile{}

func NewEmptyFile(n string) *MyFile {
	return NewFile(n, nil)
}

func NewFile(n string, c []byte) *MyFile {
	return &MyFile{
		myTreeNode: myTreeNode{name: n},
		content:    c,
	}
}

func (fi *MyFile) Accept(v TreeVisitor) {
	v.VisitFile(fi)
}

// Load loads the file contents into the file object.
func (fi *MyFile) Load(fsl *FsLoader) (err error) {
	fi.content, err = fsl.fs.ReadFile(string(fi.Path()))
	return
}

// C is the contents of the file.
func (fi *MyFile) C() []byte {
	return fi.content
}

// Equals checks for file equality
func (fi *MyFile) Equals(other *MyFile) bool {
	if fi == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	if fi.name != other.name {
		return false
	}
	if len(fi.content) != len(other.content) {
		return false
	}
	for i := range fi.content {
		if fi.content[i] != other.content[i] {
			return false
		}
	}
	return true
}
