class CodeLabelController {
    constructor(id) {
        this.id = id;
        this.el = document.getElementById('codeLabelId' + id);
        if (this.el == null) {
            console.debug("Unable to find codeLabelId = ", id)
        }
    }

    get textArea() {
        return this.el.children[0];
    }

    addCheckMark() {
        addCheckMark(this.el);
    }

    removeAllCheckMarks() {
        let numKids = this.el.children.length;
        for (let i= numKids; i > 1; i-- ) {
            this.el.removeChild(this.el.lastChild)
        }
    }

    setLabel(l) {
        this.textArea.innerText = l;
    }

    onClick(f) {
        this.el.addEventListener('click', f);
    }

    toggle() {
        if (this.isActive) {
            this.deActivate();
        } else {
            this.activate();
        }
    }

    get isActive() {
        return this.el.classList.contains('codeLabelActivated');
    }

    activate() {
        this.el.classList.remove('codeLabelDeactivated');
        this.el.classList.add('codeLabelActivated');
    }

    deActivate() {
        this.el.classList.add('codeLabelDeactivated');
        this.el.classList.remove('codeLabelActivated');
    }
}
