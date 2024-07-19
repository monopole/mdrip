class HelpBoxController {
    constructor(ntc) {
        this.navTopController = ntc
        this.style = getDocElByClass('helpBox').style;
        this.style.top = this.navTopController.height;
        this.hideIt();
    }
    hideIt() {
        this.style.height = '0px';
        this.style.removeProperty('border');
        this.style.removeProperty('border-radius');
        this.style.removeProperty('box-shadow');
    }
    showIt() {
        // if 'auto' users, then the height changes instantly - no cool transition.
        // this.style.height = 'auto';
        this.style.height = 'calc(100vh - (var(--layout-nav-bottom-height) + ' + this.navTopController.height + '))';
        this.style.border =  'solid 1px #555';
        this.style.borderRadius = '1rem';
        this.style.boxShadow = '0px 2px 2px 1px rgba(0,0,0,.3), 2px 0px 2px 1px rgba(0,0,0,.3)';
    }
    get isViz() {
        return (this.style.height !== '0px')
    }
    toggle() {
        if (this.isViz) {
            this.hideIt()
        } else {
            this.showIt()
        }
    }
}
