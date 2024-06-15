class BurgerBarsController {
    constructor() {
      this.leBurg = document.getElementById('burgerBars');
      this.onClick(() => { this.toggle(); })
    }

    hide() {
        this.leBurg.style.display = 'none';
    }

    turnOn() {
        this.leBurg.classList.add('open');
    }

    turnOff() {
        this.leBurg.classList.remove('open');
    }

    get isViz() {
        return this.leBurg.classList.contains('open');
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
