class BurgerBarsController {
    constructor() {
      this.leBurg = getDocElByClass('burgerBars');
      this.onClick(() => { this.toggle(); })
    }

    hide() {
        this.leBurg.style.display = 'none';
    }

    turnOn() {
        this.leBurg.classList.add('burgerIsAnX');
    }

    turnOff() {
        this.leBurg.classList.remove('burgerIsAnX');
    }

    get isViz() {
        return this.leBurg.classList.contains('burgerIsAnX');
    }

    toggle() {
        if (this.isViz) {
            this.turnOff();
        } else {
            this.turnOn();
        }
    }

    onClick(f) {
        this.leBurg.addEventListener('click', f);
    }
}
