// NavLeftFolderController controls a folder entry in the left nav.
// A folder entry has two states:
//  - Closed; can see its name, but not the children.
//  - Open; can see both name and children.
// Additionally, but beyond the control available here, the folder entry may
// be invisible because the encapsulating folder is closed.
class NavLeftFolderController {
    constructor(id) {
        let el = document.getElementById('navLeftFolderId' + id);
        if (el == null) {
            console.debug("Unable to find folder id = ", id)
        }
        this.children = getElByClass(el, 'navLeftFolderChildren')
        el = getElByClass(el, 'navLeftFolderName')
        el.addEventListener('click', () => {this.toggle();});
    }

    get isViz() {
        return (this.children.style.display !== 'none');
    }

    showChildren() {
        this.children.style.display = 'block';
    }

    hideChildren() {
        this.children.style.display = 'none';
    }

    toggle() {
        if (this.isViz) {
            this.hideChildren();
        } else {
            this.showChildren();
        }
    }
}
