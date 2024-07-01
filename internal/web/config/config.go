package config

const (
	// dynamicPrefix is the prefix for dynamic, rendering required
	// requests, POST requests and special requests like
	// telling the server to quit.  This distinguishes
	// such paths from paths to static content like images.
	// TODO: put up distinct ports for static vs dynamic?
	dynamicPrefix = "/_/"
)

//go:generate stringer -type=Route -linecomment
type Route int

const (
	RouteUnknown Route = iota
	// RouteJs is the GET endpoint for most of the javascript needed by the webapps.
	RouteJs // js
	// RouteCss is the GET endpoint for all the css needed by the webapps.
	RouteCss // css
	// RouteReload tells the server to reload all data from the file system.
	RouteReload // reload
	// RouteLabelsForFile is the GET endpoint for code block labels of one markdown file.
	RouteLabelsForFile // labelsForFile
	// RouteHtmlForFile is the GET endpoint for HTML of one markdown file.
	RouteHtmlForFile // htmlForFile
	// RouteRunBlock is the POST endpoint to trigger code block execution.
	RouteRunBlock // runCodeBlock
	// RouteSave is the POST endpoint to save application state.
	RouteSave // save
	// RouteImage returns an image.
	RouteImage // image
	// RouteQuit tells the server to quit.
	RouteQuit // quit
	// RouteDebug tells the server to render a debug page.
	RouteDebug // debug
	// RouteWebSocket sets up a socket.
	RouteWebSocket // debug
)

func Dynamic(r Route) string {
	return dynamicPrefix + r.String()
}

// These URL query parameter and cookie field names
// should all be unique and preferably short.
const (

	// KeyMdSessID is the param name for session ID.
	KeyMdSessID = "sid"
	// KeyIsTitleOn is the param name for is-the-title-on boolean.
	KeyIsTitleOn = "tit"
	// KeyIsNavOn is the param name for the is-the-nav-on boolean.
	KeyIsNavOn = "nav"
	// KeyMdFileIndex is the param name for the markdown file index.
	KeyMdFileIndex = "fix"
	// KeyBlockIndex is the param name for the code block index.
	KeyBlockIndex = "bix"
)
