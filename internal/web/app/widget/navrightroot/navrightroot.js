class NavRightRootController {
    constructor(appState) {
        this.appState = appState;
        this.myFileIndex = BadId;
        this.oldCodeBlockIndex = BadId;
        this.root = getDocElByClass('navRightRoot');
        this.labelController = new Array(appState.maxCodeBlocksInAFile);
        for (let i = 0; i < appState.maxCodeBlocksInAFile; i++) {
            this.labelController[i] = new CodeLabelController(i);
        }
        appState.addCodeBlockChangeReactor(this);
        appState.addFileChangeReactor(this);
        appState.addCodeBlockRunReactor(this);
        this.wireUpHandlers();
    }

    get cbIndex() {
        return this.appState.myCodeBlockIndex;
    }

    reactFileChange() {
        if (this.myFileIndex === this.appState.fileIndex) {
            return;
        }
        this.resetAllLabelControllers();
        this.myFileIndex = this.appState.fileIndex
    }

    reactCodeBlockChange() {
        if (this.oldCodeBlockIndex === this.cbIndex) {
            return;
        }
        if (this.appState.isGoodCodeBlockIndex(this.oldCodeBlockIndex)) {
            this.labelController[this.oldCodeBlockIndex].deActivate();
        }
        this.oldCodeBlockIndex = this.cbIndex
        if (!this.appState.isGoodCurrCodeBlockIndex) {
            return;
        }
        this.labelController[this.oldCodeBlockIndex].activate();
    }

    reactCodeBlockRun(index) {
        this.labelController[index].addCheckMark();
    }

    resetAllLabelControllers() {
        for (let i = 0; i < this.appState.currCodeBlocks.length; i++) {
            let c = this.labelController[i];
            c.deActivate();
            c.setLabel(this.appState.currCodeBlocks[i]);
            c.removeAllCheckMarks();
            let runCounts = this.appState.currCbRunCounts;
            for (let j = 0; j < runCounts[i]; j++) {
                c.addCheckMark();
            }
        }
        for (let i = this.appState.currCodeBlocks.length; i < this.labelController.length; i++) {
            let c = this.labelController[i];
            c.deActivate();
            c.setLabel("");
            c.removeAllCheckMarks();
        }
    }

    onClick(f) {
        this.root.addEventListener('click', f);
    }

    wireUpHandlers() {
        let me = this;
        for (let i = 0; i < this.labelController.length; i++) {
            me.labelController[i].onClick(() => {
                me.appState.setCodeBlockIndex(i);
            });
        }
        {
            let kh = function(event) {
                switch (event.key) {
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
            me.root.addEventListener('keydown', kh, false);
        }
        me.root.onmouseover = function () {
            me.root.focus();
        }
        me.root.onmouseout = function () {
            me.root.blur();
        }
    }
}
