// NavLeftFileController controls a file entry in the left nav.
// A file entry has three states:
//  - Inactive and no mouse over it.
//  - Inactive with mouse over it, ready to activate.
//  - Activated, mouse hover has no effect.
// Additionally, but beyond the control available here, the file entry may
// be invisible because the encapsulating folder is closed.
class NavLeftFileController {
    constructor(id) {
        this.id = id;
        this.el = document.getElementById('navLeftFileId' + id);
        if (this.el == null) {
            console.debug("Unable to find nav left file id = ", id)
        }
    }

    onClick(f) {
        this.el.addEventListener('click', f);
    }

    activate() {
        this.el.classList.remove('navLeftFileDeactivated');
        this.el.classList.add('navLeftFileActivated');
    }

    deActivate() {
        this.el.classList.add('navLeftFileDeactivated');
        this.el.classList.remove('navLeftFileActivated');
    }
}
