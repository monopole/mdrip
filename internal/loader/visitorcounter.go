package loader

type VisitorCounter struct {
	NumFiles   int
	NumFolders int
}

func NewVisitorCounter() *VisitorCounter {
	return &VisitorCounter{}
}

func (v *VisitorCounter) VisitTopFolder(fl *MyTopFolder) {
	// Don't count the top as a folder!
	fl.VisitChildren(v)
}

func (v *VisitorCounter) VisitFolder(fl *MyFolder) {
	v.NumFolders++
	fl.VisitChildren(v)
}

func (v *VisitorCounter) VisitFile(fi *MyFile) {
	v.NumFiles++
}

func (v *VisitorCounter) Error() error { return nil }
