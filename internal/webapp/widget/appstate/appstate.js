const StartAt = Object.freeze({
    Top: 1,
    Bottom: 0,
});
const ActivateBlock = Object.freeze({
    Yes: 1,
    No: 0,
});

class AppState {
    constructor(sessCtl, initRender) {
        this.sessionController = sessCtl;
        this.orderedPaths = initRender.OrderedPaths;
        this.myNumFolders = initRender.Facts.NumFolders;
        this.myFileIndex = initRender.Facts.InitialFileIndex;
        this.myCodeBlockIndex = initRender.Facts.InitialCodeBlockIndex;
        this.maxCodeBlocksInAFile = initRender.Facts.MaxCodeBlocksInAFile;
        this.isNavVisible = initRender.Facts.IsNavVisible;
        this.isTitleVisible = initRender.Facts.IsTitleVisible;
        this.file = {
            Html: "<p> Oops </p>",
            CodeBlockLabels: [],
            CbRunCount: [],
        };
        this.markdownRoot = document.getElementById("mdFilesRoot");
        this.fileChangeReactors = [];
        this.layoutReactors = [];
        this.codeBlockChangeReactors = [];
        this.codeBlockRunReactors = [];
    }

    report() {
        console.debug("     myFileIndex = ",this.myFileIndex);
        console.debug("myCodeBlockIndex = ",this.myCodeBlockIndex);
        console.debug("    isNavVisible = ",this.isNavVisible);
        console.debug("  isTitleVisible = ",this.isTitleVisible);
        console.debug("    myNumFolders = ",this.myNumFolders);
        console.debug("    orderedPaths = ",this.orderedPaths);
    }

    runCodeBlock() {
        let index = this.myCodeBlockIndex;
        this.sessionController.runBlock(
            this.myFileIndex, this.myCodeBlockIndex,
            () => {this.notifyCodeBlockRunReactors(index);});
    }

    focusMarkdownRoot() {
        let el = getElByClass(this.markdownRoot, "mdFilesContent");
        el.focus();
    }

    addFileChangeReactor(r) {
        this.fileChangeReactors.push(r);
    }

    addLayoutReactor(r) {
        this.layoutReactors.push(r);
    }

    toggleTitle() {
        this.isTitleVisible = !this.isTitleVisible;
        this.notifyLayoutReactors();
    }

    toggleNav() {
        this.isNavVisible = !this.isNavVisible;
        this.notifyLayoutReactors();
    }

    notifyLayoutReactors() {
        this.sessionController.save(this);
        this.layoutReactors.forEach(
            (item,i) => {item.reactLayoutChange()});
    }

    addCodeBlockChangeReactor(r) {
        this.codeBlockChangeReactors.push(r);
    }

    addCodeBlockRunReactor(r) {
        this.codeBlockRunReactors.push(r);
    }

    get fileIndex() {
        return this.myFileIndex;
    }

    isGoodFileIndex(i) {
        return (i > -1) && (i < this.numFiles);
    }

    // zero takes the application to a default "zero" state
    // like an initial session-less load.
    zero() {
        this.setFileIndex(0);
        this.setCodeBlockIndex(BadId);
        this.isNavVisible = false;
        this.isTitleVisible = true;
        this.notifyLayoutReactors();
    }

    setFileIndex(i) {
        if (!this.isGoodFileIndex(i)) {
            console.debug("bad setFileIndex: ", i);
            return;
        }
        this.myFileIndex = i;
        this.loadCurrentFile(StartAt.Top,ActivateBlock.No);
    }

    get isGoodCurrCodeBlockIndex() {
        return this.isGoodCodeBlockIndex(this.myCodeBlockIndex);
    }

    isGoodCodeBlockIndex(i) {
        return i > -1 && i < this.numCodeBlocks;
    }

    get numCodeBlocks() {
        return this.currCodeBlocks.length;
    }

    setCodeBlockIndex(i) {
        this.myCodeBlockIndex = i;
        this.notifyCodeBlockChangeReactors();
    }

    get numFiles() {
        return this.orderedPaths.length;
    }

    get numFolders() {
        return this.myNumFolders;
    }

    get prevPath() {
        if (this.myFileIndex < 1) {
            return "";
        }
        return this.orderedPaths[this.myFileIndex - 1];
    }

    get currPath() {
        return this.orderedPaths[this.myFileIndex];
    }

    get nextPath() {
        if (this.myFileIndex + 1 >= this.numFiles) {
            return "";
        }
        return this.orderedPaths[this.myFileIndex + 1];
    }

    // loadCurrentFile loads the current file and its labels,
    // and activate (or not) the topmost or bottommost code block,
    // depending on the navigation direction.
    loadCurrentFile(direction, activate) {
        this.sessionController.getFileData(
            this.fileIndex,
            (file) => {
                this.file = file;
                if (direction === StartAt.Top) {
                    if (activate === ActivateBlock.Yes) {
                        this.myCodeBlockIndex = 0;
                    } else {
                        this.myCodeBlockIndex = BadId;
                    }
                } else {
                    if (activate === ActivateBlock.Yes) {
                        this.myCodeBlockIndex = this.file.CodeBlockLabels.length - 1;
                    } else {
                        this.myCodeBlockIndex = this.file.CodeBlockLabels.length;
                    }
                }
                this.fileChangeReactors.forEach(
                    (item,i) => {item.reactFileChange()});
                this.focusMarkdownRoot();
                this.notifyCodeBlockChangeReactors();
            })
    }

    notifyCodeBlockChangeReactors() {
        this.sessionController.save(this);
        this.codeBlockChangeReactors.forEach(
            (item,i) => {item.reactCodeBlockChange()});
    }

    notifyCodeBlockRunReactors(index) {
        this.sessionController.save(this);
        this.codeBlockRunReactors.forEach(
            (item,i) => {item.reactCodeBlockRun(index)});
    }

    get currHtml() {
        return this.file.Html;
    }

    get currCodeBlocks() {
        return this.file.CodeBlockLabels;
    }

    get currCbRunCounts() {
        return this.file.CbRunCount;
    }

    goPrevFile(activate) {
        if (this.myFileIndex <= 0) {
            // Do nothing.
            return;
        }
        --(this.myFileIndex);
        this.loadCurrentFile(StartAt.Bottom, activate);
    }

    goRandomFile() {
        this.myFileIndex = randomInt(this.numFiles);
        this.loadCurrentFile(StartAt.Top, ActivateBlock.No);
    }

    goNextFile(activate) {
        if (this.myFileIndex >= this.numFiles - 1) {
            // Do nothing, not even modulo wrap.
            return;
        }
        ++(this.myFileIndex);
        this.loadCurrentFile(StartAt.Top, activate);
    }

    goRandomCodeBlock() {
        this.myFileIndex = randomInt(this.numCodeBlocks);
        this.notifyCodeBlockChangeReactors();
    }

    goPrevCodeBlock() {
        if (this.myCodeBlockIndex < 0) {
            // This means no code block is currently active.
            // We've already decremented past the legal code block.
            this.goPrevFile(ActivateBlock.Yes);
            return;
        }
        // This can go down to -1, and thus be a bad value.
        // Allows the case of zero code blocks activated.
        --(this.myCodeBlockIndex);
        this.notifyCodeBlockChangeReactors();
        //this.focusMarkdownRoot();
    }

    goNextCodeBlock() {
        if (this.myCodeBlockIndex >= this.numCodeBlocks) {
            // This means no code block is currently active.
            // We've already incremented past the legal code block.
            this.goNextFile(ActivateBlock.Yes);
            return;
        }
        // This can go up to the length of the array, and thus be a bad value.
        // Allows the case of zero code blocks activated.
        ++(this.myCodeBlockIndex);
        this.notifyCodeBlockChangeReactors();
        //this.focusMarkdownRoot();
    }
}
