function tstMakeAppState() {
    // Cannot fetch without a server, so send a full cache.
    let sc = new SessionController({{.AppState.RenderedFiles}})
    let as = new AppState(sc, {{.AppState.InitialRender}})
    tstWireArrowKeys(as);
    return as;
}

function tstWireArrowKeys(as) {
    window.addEventListener(
        'keydown', function (event) {
        if (event.defaultPrevented) {
            return;
        }
        switch (event.key) {
            case 'ArrowUp':
                as.goPrevCodeBlock();
                break;
            case 'ArrowDown':
                as.goNextCodeBlock();
                break;
            case 'ArrowLeft':
                as.goPrevFile(ActivateBlock.No);
                break;
            case 'ArrowRight':
                as.goNextFile(ActivateBlock.No);
                break;
            default:
        }
    }, false);
}

class TstReactor {
    constructor(as) {
        this.as = as
    }
    reactFileChange() {
        console.debug('TstReactor file change: ', as.fileIndex);
    }
    reactCodeBlockChange() {
        console.debug('TstReactor code block change: ', as.codeBlockIndex);
    }
}
