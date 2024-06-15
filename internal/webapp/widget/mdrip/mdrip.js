// The mdrip widget presents a markdown-based tutorial.
//
// The widget is a view into a folder tree. A folder holds markdown
// files and more folders. The markdown files are the leaves.
//
// At any given moment, exactly one <<file>> is rendered,
// and if the file has code blocks, exactly one of them
// is <<activated>> (selected for execution).
//
// Each file is numbered in depth-first order, and each code
// block on a file is numbered. So the state for the app
// is these two numbers, plus maybe some booleans indicating
// visibility of left nav, header, helpbox screen, etc.
//
//	┌─────────────────────────────────────────────────────────┐
//	│                        {title}                          │
//	│  [X]                  {current}                         │
//	│                  {prev} < ? > {next}                    │
//	├─────────────────────────────────────────────────────────┤
//	│                     {help drawer}                       │
//	├─────────────┬────────────────────────────┬──────────────┤
//	│ {file0}     │┌──────────────────────────┐│ {codeBlock0} │
//	│ {file1}     ││                          ││ {codeBlock1} │
//	│ {file2}     ││                          ││<<codeBlock2>>│
//	│ {folderA}   ││  <<contents of file3>>   ││ {codeBlock3} │
//	│  <<file3>>  ││                          ││ ...          │
//	│   {file4}   ││  This is the only        ││              │
//	│ {folderB}   ││  visible div in a column ││ Above are the│
//	│   {file5}   ││  of N divs, each holding ││ names of the │
//	│   {file6}   ││  HTML rendered from one  ││ code blocks  │
//	│   {folderC} ││  markdown file.          ││ in {file3}.  │
//	│     {file7} ││                          ││              │
//	│        ...  ││  The renderer, not a Go  ││ The name is  │
//	│             ││  template, must insert   ││ taken from   │
//	│             ││  attributes needed by    ││ the first    │
//	│             ││  the app.                ││ label on the │
//	│             ││                          ││ block.       │
//	│             ││                          ││              │
//	│             │└──────────────────────────┘│              │
//	├─────────────┴────────────────────────────┴──────────────┤
//	│                  {prev} < ? > {next}                    │
//	└─────────────────────────────────────────────────────────┘
//
class MdRipController {
    constructor(as) {
        this.appState = as;
        let tlcBottom = new TimelineController(as,{{.TimelineIdBot}});
        let tlcTop = new TimelineController(as, {{.TimelineIdTop}});
        this.ntc = new NavTopController(as, tlcTop);
        this.hbc = new HelpBoxController(this.ntc);

        tlcTop.helpButtonController.onClick(() => {
            this.hbc.toggle();
        })
        tlcBottom.helpButtonController.onClick(() => {
            this.hbc.toggle();
        })
        this.bbc = new BurgerBarsController();
        this.crc = new NavigatedContentRowController(as);
        this.mfc = new MdFilesController(as);
        let nlc = new NavLeftRootController(as);
        let nrc = new NavRightRootController(as);
        this.mkc = new MonkeyController(as, this.hbc);
        this.wireUpHandlers();
    }

    wireUpHandlers() {
        let nac = this;
        this.bbc.onClick(() => {nac.appState.toggleNav();})
        let keyHandler = function (event) {
            if (event.defaultPrevented) {
                return;
            }
            switch (event.key) {
                case 'r':
                    console.debug('reloading')
                    window.location.href = "/";
                    break;
                case 'x':
                    nac.mfc.scrollToActiveCodeBlock();
                    break;
                case '!':
                    nac.mkc.toggle();
                    break;
                case '-':
                    nac.appState.toggleTitle();
                    break;
                case 'n':  // Show left and right nav
                    nac.bbc.toggle();
                    nac.appState.toggleNav();
                    break;
                case 'Escape':
                case '/':
                case '?':
                    nac.hbc.toggle();
                    break;
                case 'a':
                case 'h':
                case 'ArrowLeft':
                    nac.appState.goPrevFile(ActivateBlock.No);
                    break;
                case 'd':
                case 'l':
                case 'ArrowRight':
                    nac.appState.goNextFile(ActivateBlock.No);
                    break;
                default:
            }
        }
        window.addEventListener('keydown', keyHandler, false);
        // TODO: on initial load, go to the current codeblock from session
        // window.setTimeout(codeBlockController.goCurrent, 700);
    }
}
