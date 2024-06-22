class SessionController {
    constructor(rf) {
        // Allow disabling of certain features for tests.
        this.enabled = false;
        this.isCodeRunning = false;
        this.isSessionSavingEnabled = false;
        // rfCache is a local cache of rendered files.
        // rfCache should be []HtmlAndLabels, i.e.
        // an array of { Html string, CodeBlockLabels, []labels CbRunCount[]int }.
        this.rfCache = rf;
    }

    enable() {
        this.enabled = true;
    }

    disable() {
        this.enabled = false;
    }

    getFileData(fileIndex, doneClosure) {
        let ans = {
            Html: "<p> Bad file index! </p>",
            CodeBlockLabels: ["ohNo"],
            CbRunCount: [0],
        }
        if (fileIndex < 0 || fileIndex >= this.rfCache.length) {
            console.debug('fileIndex out of range', fileIndex);
            doneClosure(ans);
            return;
        }
        if (this.rfCache[fileIndex] !== null) {
            console.debug('Session has cached data for fileIndex = ', fileIndex);
            doneClosure(this.rfCache[fileIndex]);
            return;
        }
        console.debug('Session calling server to get data for fileIndex = ', fileIndex);
        fetch('{{.PathGetHtmlForFile}}?{{.KeyMdFileIndex}}=' + fileIndex)
            .then((r) => {
                return r.text();
            })
            .then((r) => {
                ans.Html = r;
                return fetch('{{.PathGetLabelsForFile}}?{{.KeyMdFileIndex}}=' + fileIndex);
            })
            .then((r) => {
                return r.json();
            })
            .then((r) => {
                ans.CodeBlockLabels = r;
                ans.CbRunCount = new Array(r.length)
                for (let i = 0; i < r.length; i++) {
                    ans.CbRunCount[i] = 0;
                }
                this.rfCache[fileIndex] = ans;
                doneClosure(this.rfCache[fileIndex]);
            })
    }

    reload(doneClosure) {
        console.debug('Session calling server to reaload all data');
        fetch('{{.PathReload}}', {
            // See nearby note regarding POST.
            method: "POST",
        }).then((r) => {
            console.debug('reloaded data')
            doneClosure();
        })
    }

    // A note regarding POST.
    // In the fetch calls below we're telling the server to do a thing that
    // changes things on the server side - so we use a POST per HTTP
    // tradition.
    // But - we're sending the necessary tiny bit of data as _URL params_
    // rather than screwing around encoding and decoding a JSON body,
    // and trying to somehow map templated field names like {{.KeyBlockIndex}}
    // to JSON struct fields in both the client and server. Not worth the trouble.

    save(appState) {
        if (!this.enabled) {
            console.debug("session saving temporarily disabled; not saving state")
            return;
        }
        // TODO: DO WE WANT TO SAVE STATE SERVER SIDE?
        if (!this.isSessionSavingEnabled) {
            // console.debug("session saving hard-disabled; not saving state")
            return;
        }
        let url = '{{.PathSave}}'
            + '?{{.KeyMdFileIndex}}=' + appState.fileIndex
            + '&{{.KeyBlockIndex}}=' + appState.myCodeBlockIndex
            + '&{{.KeyIsTitleOn}}=' + appState.isTitleVisible
            + '&{{.KeyIsNavOn}}=' + appState.isNavVisible;
        fetch(url, {
            // See nearby note regarding POST.
            method: "POST",
        }).then((r) => {
            console.debug('saved session')
        })
    }

    runBlock(fileIndex, codeBlockIndex, doneClosure) {
        if (!this.enabled) {
            console.debug("session disabled; not running block")
            return;
        }
        if (this.isCodeRunning) {
            alert('busy!');
            return;
        }
        this.isCodeRunning = true;
        let me = this;
        let url = '{{.PathRunBlock}}'
            + '?{{.KeyMdFileIndex}}=' + fileIndex
            + '&{{.KeyBlockIndex}}=' + codeBlockIndex
            + '&{{.KeyMdSessID}}={{.MdSessID}}';
        fetch(url, {
            // See nearby note regarding POST.
            method: "POST",
        }).then((r) => {
            me.isCodeRunning = false;
            this.recordRunBlock(fileIndex, codeBlockIndex);
            doneClosure();
        })
    }

    recordRunBlock(fileIndex, codeBlockIndex) {
        let f = this.rfCache[fileIndex];
        if (f === null) {
            console.debug('cannot record code block run for fileIndex=', fileIndex);
            return;
        }
        if (codeBlockIndex < 0 || codeBlockIndex >= f.CbRunCount.length) {
            console.debug('cannot record code block run for codeBlockIndex=', codeBlockIndex);
            return;
        }
        f.CbRunCount[codeBlockIndex]++;
    }
}
