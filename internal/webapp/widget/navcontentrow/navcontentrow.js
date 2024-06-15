class NavLrController {
    constructor(elSp, elAct) {
        this.elSpacer = elSp;
        this.elActual = elAct;
        const bound = this.elActual
        this.elActual.onmouseover = function() {
            bound.focus();
        }
        this.elActual.onmouseout = function(){
            bound.blur();
        }
    }

    makeStyleSkinny(st) {
        // st.overflow = 'hidden';
        st.width = '0';
        st.minWidth = '0';
    }

    makeStyleWide(st) {
        // el.overflow = 'auto';
        st.width = 'var(--layout-nav-lr-width)';
        st.minWidth = 'var(--layout-nav-lr-width)';
    }

    makeStyleShorter(st) {
        st.top = "var(--layout-nav-bottom-height)";
        st.height = "calc(100vh - var(--layout-nav-bottom-height) - var(--layout-nav-bottom-height))";
    }

    makeStyleTaller(st) {
        st.top = "var(--layout-nav-top-height)";
        st.height = "calc(100vh - var(--layout-nav-top-height) - var(--layout-nav-bottom-height))";
    }

    makeTopShorter() {
        this.makeStyleShorter(this.elSpacer.style);
        this.makeStyleShorter(this.elActual.style);
    }

    makeTopTaller() {
        this.makeStyleTaller(this.elSpacer.style);
        this.makeStyleTaller(this.elActual.style);
    }

    makeSkinny() {
        this.makeStyleSkinny(this.elSpacer.style);
        this.makeStyleSkinny(this.elActual.style);
    }

    makeWide() {
        this.makeStyleWide(this.elSpacer.style);
        this.makeStyleWide(this.elActual.style);
    }
}

class NavigatedContentRowController {
    constructor(as) {
        this.appState = as;
        this.isNavViz = as.isNavVisible;
        this.isTitleViz = as.isTitleVisible;
        let el = getDocElByClass('ncrWrapper');
        this.conNavLeft = new NavLrController(
            getElByClass(el,'ncrNavSpLeft'),
             getElByClass(el,'ncrNavActLeft'))
        this.conNavRight = new NavLrController(
            getElByClass(el,'ncrNavSpRight'),
            getElByClass(el,'ncrNavActRight'))
        this.elTopSp = getElByClass(el,'ncrNavTopSp');
        this.elTopAct = getElByClass(el,'ncrNavTopAct');
        this.elCenter = getElByClass(el,'ncrContentCenter');
        as.addLayoutReactor(this);
        // If the incoming state doesn't look like the OOTB HTML layout...
        if (!(this.isTitleViz === true && this.isNavViz === false)) {
            this.reactLayoutChange()
        }
    }

    reactLayoutChange() {
        if (this.isNavViz !== this.appState.isNavVisible) {
            if (this.isNavViz) {
                this.hideNav()
            } else {
                this.showNav()
            }
            this.isNavViz = this.appState.isNavVisible
        }
        if (this.isTitleViz !== this.appState.isTitleVisible) {
            if (this.isTitleViz) {
                this.makeTopShorter()
            } else {
                this.makeTopTaller()
            }
            this.isTitleViz = this.appState.isTitleVisible
        }
    }

    hideNav() {
        this.conNavLeft.makeSkinny();
        this.conNavRight.makeSkinny();
        this.makeCenterWide();
    }

    showNav() {
        this.conNavLeft.makeWide();
        this.conNavRight.makeWide();
        this.makeCenterSkinny();
    }

    makeCenterWide() {
        this.elCenter.style.width = '100vw';
        this.elCenter.classList.remove('ncrOnShadow');
        this.elCenter.classList.add('ncrOffShadow');
    }

    makeCenterSkinny() {
        this.elCenter.style.width = 'calc(100vw - (2 * var(--layout-nav-lr-width)))';
        this.elCenter.classList.remove('ncrOffShadow');
        this.elCenter.classList.add('ncrOnShadow');
    }

    makeTopShorter() {
        this.makeStyleShorter(this.elTopSp.style);
        this.makeStyleShorter(this.elTopAct.style);
        this.conNavLeft.makeTopShorter();
        this.conNavRight.makeTopShorter();
    }

    makeTopTaller() {
        this.makeStyleTaller(this.elTopSp.style);
        this.makeStyleTaller(this.elTopAct.style);
        this.conNavLeft.makeTopTaller();
        this.conNavRight.makeTopTaller();
    }

    makeStyleShorter(st) {
        st.height = "var(--layout-nav-bottom-height)";
    }

    makeStyleTaller(st) {
        st.height = "var(--layout-nav-top-height)";
    }
}
