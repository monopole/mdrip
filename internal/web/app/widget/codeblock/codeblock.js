// CodeBlockController manages one codeBlock.
// It's similar in scope to CodeLabelController.
// Likewise, MdFilesController is similar in scope to NavRightRootController.
//
// A codeBlock has three states:
//
// * DeActivated - unselected, idle, possibly annotated with past executions.
//               Only goes to "Activated".
// * Activated - selected and ready to execute.
//               Goes to either of the other two states.
// * Executing - Waiting for server to signal completion.
//               Only goes to deActivated, and it adds a checkmark or something
//               to indicate that it ran.
//
//  In practice, since only one block can run at a time, the entire app is either
//  running the active codeBlock or not.
class CodeBlockController {
    constructor(id) {
        this.id = id;
        this.el = null;
        this.onClickFunctions = [];
    }

    // No method on this controller can be called before
    // at least one call to "reset".
    reset() {
        if (this.el != null) {
            this.deActivate();
            this.onClickFunctions.forEach(
                (f) => {this.el.removeEventListener('click', f)})
        }
        this.onClickFunctions = [];
        this.el = document.getElementById('codeBlockId' + this.id);
        let me = this;
        this.addOnClick(()=>{
            me.attemptCopyToBuffer();
        })
    }

    addOnClick(f) {
        this.onClickFunctions.push(f);
        this.el.addEventListener('click', f);
    }

    get controlBar() {
        return this.el.children[0];
    }

    get prompt() {
        return this.el.children[1];
    }

    get codeArea() {
        return this.el.children[2];
    }

    get isActive() {
        return this.prompt.style.display === 'inline-block';
    }

    toggle() {
        if (this.isActive) {
            this.deActivate();
        } else {
            this.activate();
        }
    }

    deActivate() {
        this.prompt.style.display = 'none';
        this.codeArea.style.boxShadow = '';
        this.codeArea.style.color = 'var(--color-code-inactive)';
    }


    activate() {
        this.prompt.style.display = 'inline-block';
        // box-shadow is a rectangle shape - a shadow - behind the object.
        //     If no offset, it's invisible.
        // box-shadow: 'color [inset] offset-x offset-y blur-radius spread-radius'
        // offset-x: if positive, the shadow is on the right
        // offset-y: if positive, the shadow is on the bottom
        // blur-radius: bigger and lighter, cannot be negative.
        //              If not specified, it's zero meaning sharp edge.
        // spread-radius: positive values make the shadow grow,
        //                negative values make is shrink.
        // inset - make the shadow inside the box.
        // The "alpha" value of the color is the opacity.
        //     0==transparent, 1==opaque
        this.codeArea.style.boxShadow = 'var(--color-code-shadow) 3px 3px 5px';
        this.codeArea.style.color = 'var(--color-code-active)';
        this.scrollIntoView();
    }

    scrollIntoView() {
        // console.debug("scrolling code block ", this.el.getAttribute("id"), " into view")
        // console.debug("document.activeElement = ", document.activeElement.getAttribute("class"))
        this.el.scrollIntoView(
            {behavior: 'smooth', block: 'center', inline: 'nearest'});
    }

    addCheckMark() {
        addCheckMark(this.controlBar);
    }

    // ---------------------------------------------
    // The following involves copying code from the
    // codeBlock into the copy/paste buffer.
    // ---------------------------------------------

    attemptCopyToBuffer() {
        // The [3] seems fragile, but it works. See codeBlock.html
        // childNodes[1] is codeBlockControl
        // childNodes[3] is codeBlockArea
        let text = this.codeArea.firstChild.textContent
        if (!navigator.clipboard) {
            this.oldAttemptCopyToBuffer(text);
            return;
        }
        navigator.clipboard.writeText(text).then(function() {
            // console.log('Async: Copying to clipboard was successful!');
        }, function(err) {
            console.error('Oops1, unable to copy to paste buffer', err);
        });
    }

    // https://stackoverflow.com/questions/400212
    oldAttemptCopyToBuffer(text) {
        let tA = document.createElement('textarea');
        this.hideCopyPasteTextArea(tA.style);
        tA.value = text;
        document.body.appendChild(tA);
        tA.select();
        try {
            let successful = document.execCommand('copy');
            let msg = successful ? 'successful' : 'unsuccessful';
            console.log('Fallback: Copying text command was ' + msg);
        } catch (err) {
            console.error('Oops2, unable to copy to paste buffer');
        }
        document.body.removeChild(tA);
    }

    hideCopyPasteTextArea(s) {
        s.position = 'fixed';
        s.top = 0;
        s.left = 0;
        s.width = '2em';
        s.height = '2em';
        s.padding = 0;
        s.border = 'none';
        s.outline = 'none';
        s.boxShadow = 'none';
        s.background = 'transparent';
    }
}
