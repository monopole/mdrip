package subshell

import (
	"strings"
	"testing"
	"time"

	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/program"
	"github.com/monopole/mdrip/scanner"
)

const timeout = 2 * time.Second

func makeBlock(code string) *program.BlockPgm {
	return program.NewBlockPgm(code)
}

func newTutorial(b []*model.BlockTut) model.Tutorial {
	return model.NewLessonTut("iamafilename", b)
}

func emptyTutorial() model.Tutorial {
	return newTutorial([]*model.BlockTut{})
}

func TestRunnerWithNothing(t *testing.T) {
	if NewSubshell(
		timeout,
		program.NewProgramFromTutorial(base.WildCardLabel, emptyTutorial())).Run().Problem() != nil {
		t.Fail()
	}
}

func doIt(blocks []*program.BlockPgm) *RunResult {
	lesson := program.NewLessonPgm(base.FilePath("foo"), blocks)
	p := program.NewProgram([]*program.LessonPgm{lesson})
	return NewSubshell(timeout, p).Run()
}

func TestRunnerWithGoodStuff(t *testing.T) {
	blocks := []*program.BlockPgm{
		makeBlock("echo kale\ndate\n"),
		makeBlock("echo beans\necho apple\n"),
		makeBlock("echo hasta\necho la vista\n")}
	result := doIt(blocks)
	if result.Problem() != nil {
		t.Fail()
	}
}

func checkFail(t *testing.T, got, want *RunResult) {
	if got.Problem() == nil {
		t.Fail()
	}
	if got.Index() != want.Index() {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "file", got.Index(), want.Index())
	}
	if !strings.Contains(got.Message(), want.Message()) {
		t.Errorf("%s got\n\t\"%v\"\nwant\n\t%v", "message", got.Message(), want.Message())
	}
}

func TestStartWithABadCommand(t *testing.T) {
	want := NoCommandsRunResult(
		NewFailureOutput("dunno"),
		"fileNameTestStartWithABadCommand",
		0,
		"line 1: notagoodcommand: command not found")

	blocks := []*program.BlockPgm{
		makeBlock("notagoodcommand\ndate\n"),
		makeBlock("echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}

func TestBadCommandInTheMiddle(t *testing.T) {
	want := NoCommandsRunResult(
		NewFailureOutput("dunno"),
		"fileNameTestBadCommandInTheMiddle",
		2,
		"line 9: lochNessMonster: command not found")

	blocks := []*program.BlockPgm{
		makeBlock("echo tofu\ndate\n"),
		makeBlock("echo beans\necho kale\n"),
		makeBlock("lochNessMonster\n"),
		makeBlock("echo hasta\necho la vista\n")}

	checkFail(t, doIt(blocks), want)
}

func TestTimeOut(t *testing.T) {
	want := NoCommandsRunResult(
		NewFailureOutput("dunno"),
		"fileNameTestTimeOut",
		0,
		scanner.MsgTimeout)

	// Insert this sleep in a command block.
	// Arrange to sleep for two seconds longer than the timeout.
	sleep := timeout + (2 * time.Second)

	blocks := []*program.BlockPgm{
		makeBlock("date\nsleep " + sleep.String() + "\necho kale"),
		makeBlock("echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}
