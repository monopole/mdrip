// MdFilesController shows and hides markdown file contents.
class MdFilesController {
    constructor(appState) {
        this.appState = appState;
        this.root = document.getElementById("mdFilesRoot");
        this.myFileIndex = BadId;
        this.oldCodeBlockIndex = BadId;
        this.cbControllers = [];
        this.cbControllers = new Array(appState.maxCodeBlocksInAFile);
        for (let i = 0; i < appState.maxCodeBlocksInAFile; i++) {
            // This constructor must not attempt to find divs on
            // the page, as the html has not been loaded yet.
            this.cbControllers[i] = new CodeBlockController(i);
        }
        appState.addFileChangeReactor(this);
        appState.addCodeBlockChangeReactor(this);
        appState.addCodeBlockRunReactor(this);
    }

    makeContentDiv() {
        let el = document.createElement('div');
        el.innerHTML = this.appState.currHtml;
        el.setAttribute("class", "mdFilesContent")
        // Without the tabindex attribute, key events will not be captured.
        // https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/tabindex
        el.setAttribute("tabindex", "-1")
        return el;
    }

    reactFileChange() {
        if (this.myFileIndex === this.appState.fileIndex) {
            return;
        }
        this.myFileIndex = this.appState.fileIndex

        let newDiv = this.makeContentDiv();
        this.root.replaceChild(newDiv, this.root.firstElementChild);
        newDiv.focus();
        this.wireUpHandlers(newDiv);
        this.resetAllCodeBlocks();
        this.updateUrl();

        // This odd recursive func causes the reading area to scroll up
        // to the top of the new file.  Without it, file changes leave one
        // at the same point as in the previous file, rather than at the top.
        let smoothlyScrollToTop = function() {
            let currentScroll =
                document.documentElement.scrollTop || document.body.scrollTop;
            if (currentScroll > 0) {
                window.requestAnimationFrame(smoothlyScrollToTop);
                window.scrollTo(0,currentScroll - (currentScroll/5));
            }
        }
        smoothlyScrollToTop();
    }

    get cbIndex() {
        return this.appState.myCodeBlockIndex;
    }

    runActiveCodeBlock() {
        if (!this.appState.isGoodCurrCodeBlockIndex) {
            console.debug('No active code block.');
            return;
        }
        this.appState.runCodeBlock()
    }

    reactCodeBlockRun(index) {
        this.cbControllers[index].addCheckMark();
    }

    scrollToActiveCodeBlock() {
        if (this.appState.isGoodCurrCodeBlockIndex) {
            this.cbControllers[this.cbIndex].scrollIntoView()
        }
    }

    // TODO: get browser back/forward buttons to work.
    //   The url changes when one hits the back/forward buttons,
    //   but the page content doesn't change.
    updateUrl() {
        if (window.location.origin.startsWith("file://")) {
            // This means one is debugging a local static rendering.
            return;
        }
        let path = this.appState.currPath
        if (history.pushState) {
            window.history.pushState(
                "not using data yet", "someTitle", "/" + path);
        } else {
            document.location.href = path;
        }
    }

    resetAllCodeBlocks() {
        let me = this;
        for (let i = 0; i < this.appState.currCodeBlocks.length; i++) {
            let cbc = this.cbControllers[i];
            cbc.reset();
            cbc.addOnClick(()=>{
                me.appState.setCodeBlockIndex(i);
            });
            let runCounts = this.appState.currCbRunCounts;
            for (let j = 0; j < runCounts[i]; j++) {
                cbc.addCheckMark();
            }
        }
    }

    reactCodeBlockChange() {
        // if (this.oldCodeBlockIndex === this.cbIndex) {
        //     return;
        // }
        if (this.appState.isGoodCodeBlockIndex(this.oldCodeBlockIndex)) {
            this.cbControllers[this.oldCodeBlockIndex].deActivate();
        }
        this.oldCodeBlockIndex = this.cbIndex
        if (!this.appState.isGoodCurrCodeBlockIndex) {
            return;
        }
        this.cbControllers[this.oldCodeBlockIndex].activate();
    }

    wireUpHandlers(el) {
        const me = this;
        {
            let kh = function(event) {
                switch (event.key) {
                    case 'Enter':
                        event.preventDefault();
                        me.runActiveCodeBlock();
                        me.appState.goNextCodeBlock();
                        break;
                    case 'w':
                    case 'k':
                    case 'ArrowUp':
                        event.preventDefault();
                        me.appState.goPrevCodeBlock();
                        break;
                    case 'j':
                    case 's':
                    case 'ArrowDown':
                        event.preventDefault();
                        me.appState.goNextCodeBlock();
                        break;
                    default:
                }
            }
            el.addEventListener('keydown', kh, false);
        }
        // If this is used, the content seems to autoscroll
        // to the "top", making the top part hidden under the navtop.
        // THIS SUCKS!
        // el.onmouseover = function () {
        //     el.focus();
        // }
        // el.onmouseout = function () {
        //     el.blur();
        // }
    }
}
