package subshell

import (
	"strings"
	"testing"
	"time"

	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/scanner"
)

const timeout = 2 * time.Second

var labels = []model.Label{model.Label("foo"), model.Label("bar")}

func makeCommandBlock(labels []model.Label, code string) *model.CommandBlock {
	return model.NewCommandBlock(labels, []byte{}, model.OpaqueCode(code))
}

func newTutorial(b []*model.CommandBlock) model.Tutorial {
	return model.NewLesson("iamafilename", b)
}

func emptyTutorial() model.Tutorial {
	return newTutorial([]*model.CommandBlock{})
}

func TestRunnerWithNothing(t *testing.T) {
	if NewSubshell(
		timeout,
		model.NewProgramFromTutorial(model.AnyLabel, emptyTutorial())).Run().Problem() != nil {
		t.Fail()
	}
}

func doIt(blocks []*model.CommandBlock) *RunResult {
	p := model.NewProgramFromTutorial(model.AnyLabel, newTutorial(blocks))
	return NewSubshell(timeout, p).Run()
}

func TestRunnerWithGoodStuff(t *testing.T) {
	blocks := []*model.CommandBlock{
		makeCommandBlock(labels, "echo kale\ndate\n"),
		makeCommandBlock(labels, "echo beans\necho apple\n"),
		makeCommandBlock(labels, "echo hasta\necho la vista\n")}
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

	blocks := []*model.CommandBlock{
		makeCommandBlock(labels, "notagoodcommand\ndate\n"),
		makeCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}

func TestBadCommandInTheMiddle(t *testing.T) {
	want := NoCommandsRunResult(
		NewFailureOutput("dunno"),
		"fileNameTestBadCommandInTheMiddle",
		2,
		"line 9: lochNessMonster: command not found")

	blocks := []*model.CommandBlock{
		makeCommandBlock(labels, "echo tofu\ndate\n"),
		makeCommandBlock(labels, "echo beans\necho kale\n"),
		makeCommandBlock(labels, "lochNessMonster\n"),
		makeCommandBlock(labels, "echo hasta\necho la vista\n")}

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

	blocks := []*model.CommandBlock{
		makeCommandBlock(labels, "date\nsleep "+sleep.String()+"\necho kale"),
		makeCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}
