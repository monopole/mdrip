package util

import (
	"github.com/monopole/mdrip/model"
	"strconv"
	"strings"
	"testing"
	"time"
)

var noLabels []model.Label = []model.Label{}
var labels = []model.Label{model.Label("foo"), model.Label("bar")}
var emptyCommandBlock *model.CommandBlock = model.NewCommandBlock(noLabels, "")

const timeoutSeconds = 1

func TestRunnerWithNothing(t *testing.T) {
	if RunInSubShell(model.NewProgram(), timeoutSeconds*time.Second).Problem() != nil {
		t.Fail()
	}
}

func doIt(blocks []*model.CommandBlock) *model.ScriptResult {
	p := model.NewProgram().Add(model.NewScriptBucket("iAmFileName", blocks))
	return RunInSubShell(p, timeoutSeconds*time.Second)
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

func checkFail(t *testing.T, got, want *model.ScriptResult) {
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

	want := model.NoCommandsScriptResult(
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

	want := model.NoCommandsScriptResult(
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
	want := model.NoCommandsScriptResult(
		model.NewFailureOutput("dunno"),
		"fileNameTestTimeOut",
		0,
		MsgTimeout)

	// Go to sleep for twice the length of the timeout.
	blocks := []*model.CommandBlock{
		model.NewCommandBlock(labels, "date\nsleep "+strconv.Itoa(timeoutSeconds+2)+"\necho kale"),
		model.NewCommandBlock(labels, "echo beans\necho cheese\n")}
	checkFail(t, doIt(blocks), want)
}
