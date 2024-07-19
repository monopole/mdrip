class NavLeftRootController {
    constructor(appState) {
        this.appState = appState;
        this.myFileIndex = BadId;
        this.root = getDocElByClass('navLeftRoot');
        this.folderController = new Array(appState.numFolders);
        this.fileController = new Array(appState.numFiles);
        for (let i = 0; i < appState.numFolders; i++) {
            this.folderController[i] = new NavLeftFolderController(i);
        }
        for (let i = 0; i < appState.numFiles; i++) {
            this.fileController[i] = new NavLeftFileController(i);
        }
        appState.addFileChangeReactor(this);
        this.wireUpHandlers();
    }

    reactFileChange() {
        if (this.myFileIndex === this.appState.fileIndex) {
            return;
        }
        if (this.myFileIndex !== BadId) {
            this.fileController[this.myFileIndex].deActivate();
        }
        this.myFileIndex = this.appState.fileIndex
        this.fileController[this.myFileIndex].activate();
    }

    onClick(f) {
        this.root.addEventListener('click', f);
    }

    wireUpHandlers() {
        let me = this;
        for (let i = 0; i < this.fileController.length; i++) {
            me.fileController[i].onClick(() => {
                me.appState.setFileIndex(i);
            });
        }
        {
            let kh = function(event) {
                switch (event.key) {
                    case 'w':
                    case 'k':
                    case 'ArrowUp':
                        event.preventDefault();
                        me.appState.goPrevFile(ActivateBlock.No);
                        break;
                    case 'j':
                    case 's':
                    case 'ArrowDown':
                        event.preventDefault();
                        me.appState.goNextFile(ActivateBlock.No);
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
