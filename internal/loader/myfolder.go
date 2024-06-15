package loader

// MyFolder is a named group of files and folders.
type MyFolder struct {
	myTreeNode
	files []*MyFile
	dirs  []*MyFolder
}

var _ MyTreeNode = &MyFolder{}

func NewFolder(n string) *MyFolder {
	return &MyFolder{myTreeNode: myTreeNode{name: n}}
}

func (fl *MyFolder) Accept(v TreeVisitor) {
	v.VisitFolder(fl)
}

func (fl *MyFolder) VisitChildren(v TreeVisitor) {
	fl.visitFiles(v)
	fl.visitFolders(v)
}

func (fl *MyFolder) visitFiles(v TreeVisitor) {
	for i := range fl.files {
		fl.files[i].Accept(v)
	}
}

func (fl *MyFolder) visitFolders(v TreeVisitor) {
	for i := range fl.dirs {
		fl.dirs[i].Accept(v)
	}
}

func (fl *MyFolder) NumFiles() int {
	if fl == nil {
		return 0
	}
	return len(fl.files)
}

func (fl *MyFolder) NumFolders() int {
	if fl == nil {
		return 0
	}
	return len(fl.dirs)
}

func (fl *MyFolder) AddFile(file *MyFile) *MyFolder {
	file.parent = fl
	fl.files = append(fl.files, file)
	return fl
}

func (fl *MyFolder) AddFolder(folder *MyFolder) *MyFolder {
	folder.parent = fl
	fl.dirs = append(fl.dirs, folder)
	return fl
}
func (fl *MyFolder) IsEmpty() bool {
	if fl == nil {
		return true
	}
	return len(fl.files) == 0 && len(fl.dirs) == 0
}

func (fl *MyFolder) Equals(other *MyFolder) bool {
	if fl == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	if fl.name != other.name {
		return false
	}
	if !EqualFileSlice(fl.files, other.files) {
		return false
	}
	return EqualFolderSlice(fl.dirs, other.dirs)
}

func EqualFileSlice(s1 []*MyFile, s2 []*MyFile) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if !s1[i].Equals(s2[i]) {
			return false
		}
	}
	return true
}

func EqualFolderSlice(s1 []*MyFolder, s2 []*MyFolder) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if !s1[i].Equals(s2[i]) {
			return false
		}
	}
	return true
}

func (fl *MyFolder) HasFile(name string) bool {
	for _, fi := range fl.files {
		if fi.Name() == name {
			return true
		}
	}
	return false
}
