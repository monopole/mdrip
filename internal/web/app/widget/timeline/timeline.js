class TimelineController {
    constructor(as, id) {
        this.appState = as;
        this.elRow = document.getElementById('timelineId' + id);

        this.helpButtonController = new HelpButtonController(
            getElByClass(this.elRow, 'helpButton'));

        this.tlPrev = getElByClass(this.elRow, 'timelinePrev');
        this.titlePrev = getElByClass(this.tlPrev, 'timelineTitlePrev');
        this.arrowPrev = getElByClass(this.tlPrev, 'timelinePointer');

        this.tlNext = getElByClass(this.elRow, 'timelineNext');
        this.titleNext = getElByClass(this.tlNext, 'timelineTitleNext');
        this.arrowNext = getElByClass(this.tlNext, 'timelinePointer');

        as.addFileChangeReactor(this);
        this.wireUpHandlers();
    }

    reactFileChange() {
        this.setText(this.appState.prevPath, this.appState.nextPath);
    }

    setText(p, n) {
        this.titlePrev.innerHTML = p;
        if (p === "") {
            this.arrowPrev.innerHTML = "";
        } else {
            this.arrowPrev.innerHTML = "&lt;";
        }
        this.titleNext.innerHTML = n;
        if (n === "") {
            this.arrowNext.innerHTML = "";
        } else {
            this.arrowNext.innerHTML = "&gt;";
        }
    }

    wireUpHandlers(f) {
        this.tlPrev.addEventListener(
            'click', () => {
                this.appState.goPrevFile(ActivateBlock.No)
            });
        this.tlNext.addEventListener(
            'click', () => {
                this.appState.goNextFile(ActivateBlock.No)
            });
    }

    get isViz() {
        return (this.elRow.style.display !== 'none');
    }

    hideIt() {
        this.elRow.style.display = 'none';
    }

    showIt() {
        this.elRow.style.display = 'flex';
    }

    toggle() {
        if (this.isViz) {
            this.hideIt()
        } else {
            this.showIt()
        }
    }
}
