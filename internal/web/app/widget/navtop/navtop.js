class NavTopController {
    constructor(as, tlc) {
        this.appState = as;
        this.isTitleViz = as.isTitleVisible;
        this.timelineController = tlc;
        this.styleHeader = document.getElementById('header').style;
        this.styleTitleDoc = getDocElByClass('nvtTitleDoc').style;
        this.styleTitleCurr = getDocElByClass('nvtTitleCurr');
        this.setHeight('var(--layout-nav-top-height)');
        as.addFileChangeReactor(this);
        as.addLayoutReactor(this);
        // If the incoming state doesn't look like the OOTB HTML layout...
        if (this.isTitleViz !== true) {
            this.reactLayoutChange()
        }
    }

    reactLayoutChange() {
        if (this.isTitleViz !== this.appState.isTitleVisible) {
            if (this.isTitleViz) {
                this.beShorter()
            } else {
                this.beTaller()
            }
            this.isTitleViz = this.appState.isTitleVisible
        }
    }

    reactFileChange() {
        this.styleTitleCurr.innerHTML = this.appState.currPath;
    }

    get height() {
        return this.styleHeader.height;
    }

    setHeight(h) {
        this.styleHeader.height = h;
    }

    beShorter() {
        this.setHeight('var(--layout-nav-bottom-height)');
        this.styleTitleDoc.display = 'none';
        this.timelineController.hideIt();
    }

    beTaller() {
        this.setHeight('var(--layout-nav-top-height)');
        this.styleTitleDoc.display = 'block';
        this.timelineController.showIt();
    }
}
