package util

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

var emptyArray []string = []string{}
var emptyCommandBlock *CommandBlock = &CommandBlock{emptyArray, ""}

const timeoutSeconds = 1

func TestRunnerWithNothing(t *testing.T) {
	if RunInSubShell([]*ScriptBucket{}, timeoutSeconds*time.Second).problem != nil {
		t.Fail()
	}
}

func doIt(blocks []*CommandBlock) *ScriptResult {
	return RunInSubShell([]*ScriptBucket{&ScriptBucket{"iAmFileName", blocks}}, timeoutSeconds*time.Second)
}

func TestRunnerWithGoodStuff(t *testing.T) {
	labels := []string{"foo", "bar"}
	blocks := []*CommandBlock{
		&CommandBlock{labels, "echo kale\ndate\n"},
		&CommandBlock{labels, "echo beans\necho apple\n"},
		&CommandBlock{labels, "echo hasta\necho la vista\n"}}
	result := doIt(blocks)
	if result.problem != nil {
		t.Fail()
	}
}

func checkFail(t *testing.T, got, want *ScriptResult) {
	if got.problem == nil {
		t.Fail()
	}
	if got.index != want.index {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "script", got.index, want.index)
	}
	if !strings.Contains(got.message, want.message) {
		t.Errorf("%s got\n\t%v\nwant\n\t%v", "message", got.message, want.message)
	}
}

func TestStartWithABadCommand(t *testing.T) {
	want := &ScriptResult{
		blockOutput{false, "dunno"},
		"fileNameTestStartWithABadCommand",
		0,
		emptyCommandBlock,
		nil,
		"line 1: notagoodcommand: command not found"}

	labels := []string{"foo", "bar"}
	blocks := []*CommandBlock{
		&CommandBlock{labels, "notagoodcommand\ndate\n"},
		&CommandBlock{labels, "echo beans\necho cheese\n"}}
	checkFail(t, doIt(blocks), want)
}

func TestBadCommandInTheMiddle(t *testing.T) {
	want := &ScriptResult{
		blockOutput{false, "dunno"},
		"fileNameTestBadCommandInTheMiddle",
		2,
		emptyCommandBlock,
		nil,
		"line 9: lochNessMonster: command not found"}

	labels := []string{"foo", "bar"}

	blocks := []*CommandBlock{
		&CommandBlock{labels, "echo tofu\ndate\n"},
		&CommandBlock{labels, "echo beans\necho kale\n"},
		&CommandBlock{labels, "lochNessMonster\n"},
		&CommandBlock{labels, "echo hasta\necho la vista\n"}}

	checkFail(t, doIt(blocks), want)
}

func TestTimeOut(t *testing.T) {
	want := &ScriptResult{
		blockOutput{false, "dunno"},
		"fileNameTestTimeOut",
		0,
		emptyCommandBlock,
		nil,
		MsgTimeout}

	labels := []string{"foo", "bar"}
	// Go to sleep for twice the length of the timeout.
	blocks := []*CommandBlock{
		&CommandBlock{labels, "date\nsleep " + strconv.Itoa(timeoutSeconds+2) + "\necho kale"},
		&CommandBlock{labels, "echo beans\necho cheese\n"}}
	checkFail(t, doIt(blocks), want)
}
