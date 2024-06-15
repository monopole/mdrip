// MonkeyController does ¯\_(ツ)_/¯
// Helps confirm CSS and JS robustness.
class MonkeyController {
    constructor(as, helpController) {
        this.appState = as;
        this.helpController = helpController;
        this.monkeyActive = false;
        this.interval = null;
    }

    toggle() {
        if (this.monkeyActive) {
            this.stopMonkey();
        } else {
            this.startMonkey();
        }
    }

    startMonkey() {
        this.appState.sessionController.disable()

        // monkey should allow transitions to finish; add a delay.
        const monkeyPause = 50; // ms

        this.interval = window.setInterval(
            () => {
                this.doSomething()
            },
            parseInt('{{.TransitionSpeedMs}}', 10) + monkeyPause);

        this.monkeyActive = true;
    }

    stopMonkey() {
        window.clearInterval(this.interval);
        this.interval = null;
        this.monkeyActive = false;
        this.reset();
        this.appState.sessionController.enable()
    }

    doSomething() {
        if (this.helpController.isViz) {
            this.helpController.toggle();
            return;
        }
        switch (randomInt(3)) {
            case 0:
                this.appState.goRandomFile();
                break;
            case 1:
                this.appState.goRandomCodeBlock();
                break;
            default:
                switch (randomInt(3)) {
                    case 0:
                        this.appState.toggleTitle();
                        break;
                    case 1:
                        this.appState.toggleNav();
                        break;
                    default:
                        this.helpController.toggle();
                        break;
                }
                break;
        }
    }

    reset() {
        if (this.helpController.isViz) {
            this.helpController.hideIt();
        }
        this.appState.zero();
    }
}
