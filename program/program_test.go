package program

import (
	"strings"
	"testing"
	"time"

	"github.com/monopole/mdrip/scanner"
	"github.com/monopole/mdrip/model"
)

const timeout = 2 * time.Second

var noLabels []model.Label = []model.Label{}
var labels = []model.Label{model.Label("foo"), model.Label("bar")}

func TestRunnerWithNothing(t *testing.T) {
	if NewProgram(timeout, labels[0], []model.FileName{}).RunInSubShell().Problem() != nil {
		t.Fail()
	}
}

func doIt(blocks []*model.CommandBlock) *model.RunResult {
	p := NewProgram(timeout, labels[0], []model.FileName{}).Add(model.NewScript("iAmFileName", blocks))
	return p.RunInSubShell()
}

func TestRunnerWithGoodStuff(t *testing.T) {
	blocks := []*model.CommandBlock{
		model.NewCommandBlock(labels, "echo kale\ndate\n"),
		model.NewCommandBlock(labels, "echo beans\necho apple\n"),
		model.NewCommandBlock(labels, "echo hasta\necho la vista\n")}
	result := doIt(blocks)
	if result.Problem() != nil {
		t.Fail()
	}
}

func checkFail(t *testing.T, got, want *model.RunResult) {
	if got.Problem() == nil {
		t.Fail()
	}
	if got.Index() != want.Index() {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "script", got.Index(), want.Index())
	}
	if !strings.Contains(got.Message(), want.Message()) {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "message", got.Message(), want.Message())
	}
}

func TestStartWithABadCommand(t *testing.T) {

	want := model.NoCommandsRunResult(
		model.NewFailureOutput("dunno"),
		"fileNameTestStartWithABadCommand",
		0,
		"line 1: notagoodcommand: command not found")

	blocks := []*model.CommandBlock{
		model.NewCommandBlock(labels, "notagoodcommand\ndate\n"),
		model.NewCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}

func TestBadCommandInTheMiddle(t *testing.T) {

	want := model.NoCommandsRunResult(
		model.NewFailureOutput("dunno"),
		"fileNameTestBadCommandInTheMiddle",
		2,
		"line 9: lochNessMonster: command not found")

	blocks := []*model.CommandBlock{
		model.NewCommandBlock(labels, "echo tofu\ndate\n"),
		model.NewCommandBlock(labels, "echo beans\necho kale\n"),
		model.NewCommandBlock(labels, "lochNessMonster\n"),
		model.NewCommandBlock(labels, "echo hasta\necho la vista\n")}

	checkFail(t, doIt(blocks), want)
}

func TestTimeOut(t *testing.T) {
	want := model.NoCommandsRunResult(
		model.NewFailureOutput("dunno"),
		"fileNameTestTimeOut",
		0,
		scanner.MsgTimeout)

	// Insert this sleep in a command block.
	// Arrange to sleep for two seconds longer than the timeout.
	sleep := timeout + (2 * time.Second)

	blocks := []*model.CommandBlock{
		model.NewCommandBlock(labels, "date\nsleep "+sleep.String()+"\necho kale"),
		model.NewCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}
