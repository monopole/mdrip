class HelpButtonController {
    constructor(el) {
        this.butt = el;
    }
    onClick(f) {
        this.butt.addEventListener('click', f);
    }
}
