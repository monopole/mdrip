package loader

// MyTopFolder is a named group of files and folders.
// It behaves slightly (but critically) different from a plain folder
// in that it never has a name.
type MyTopFolder struct {
	MyFolder
}

var _ MyTreeNode = &MyTopFolder{}

func NewTopFolder(fld *MyFolder) *MyTopFolder {
	tf := &MyTopFolder{MyFolder: *fld}
	tf.name = ""
	for _, c := range tf.files {
		c.parent = tf
	}
	for _, c := range tf.dirs {
		c.parent = tf
	}
	return tf
}

func (fl *MyTopFolder) Accept(v TreeVisitor) {
	v.VisitTopFolder(fl)
}

func (fl *MyTopFolder) AddFile(file *MyFile) *MyTopFolder {
	file.parent = fl
	fl.files = append(fl.files, file)
	return fl
}

func (fl *MyTopFolder) AddFolder(folder *MyFolder) *MyTopFolder {
	folder.parent = fl
	fl.dirs = append(fl.dirs, folder)
	return fl
}

func (fl *MyTopFolder) Equals(other *MyTopFolder) bool {
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
