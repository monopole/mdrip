package loader

// TreeVisitor has the ability to visit the items specified in its methods.
type TreeVisitor interface {
	VisitTopFolder(*MyTopFolder)
	VisitFolder(*MyFolder)
	VisitFile(*MyFile)
	// Error returns the visitation error, if any.
	Error() error
}
