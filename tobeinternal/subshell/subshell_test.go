package subshell

import (
	"strings"
	"testing"
	"time"

	"github.com/monopole/mdrip/tobeinternal/base"
	program2 "github.com/monopole/mdrip/tobeinternal/program"
	"github.com/monopole/mdrip/tobeinternal/scanner"
)

// To run this test with logging:
// cd github.com/monopole/mdrip
// go test . --alsologtostderr --vmodule subshell=2 --stderrthreshold INFO

const (
	timeout = 1 * time.Second
	// Arrange for a sleep that is longer than the timeout.
	sleep = timeout + (1 * time.Second)
)

func makeBlock(code string) *program2.BlockPgm {
	return program2.NewBlockPgm(code)
}

func doIt(lines []string) *RunResult {
	var pgm []*program2.BlockPgm
	for _, l := range lines {
		pgm = append(pgm, makeBlock(l))
	}
	lesson := program2.NewLessonPgm(base.FilePath("arbitraryPath"), pgm)
	p := program2.NewProgram([]*program2.LessonPgm{lesson})
	return NewSubshell(timeout, p).Run()
}

func TestRunnerWithNothing(t *testing.T) {
	result := doIt([]string{})
	if result.Error() != nil {
		t.Errorf("Expected no error, got %v", result.Error())
	}
}

func TestRunnerWithGoodStuff(t *testing.T) {
	result := doIt([]string{
		"echo kale\ndate\n",
		"echo beans\necho apple\n",
		"echo hasta\necho la vista\n",
	})
	if result.Error() != nil {
		t.Fail()
	}
	if !result.Completed() {
		t.Fail()
	}
}

func checkFail(t *testing.T, got *RunResult, wantIndex int, wantErr string) {
	if got.Error() == nil {
		t.Errorf("expected an error, but no error")
	}
	if got.Index() != wantIndex {
		t.Errorf("got index %v, want index %v", got.Index(), wantIndex)
	}
	if !strings.Contains(got.StdErr(), wantErr) {
		t.Errorf("got stderr\n\t\"%v\"\nwant stderr\n\t%v", got.StdErr(), wantErr)
	}
}

func TestFalse(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"echo beans\necho lemon\n",
			"/bin/false\necho kale",
			"echo tofu\ndate\n",
		}),
		1,
		"")
}

func TestBadCommandAtStart(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"lochNessMonster\ndate\n",
			"echo beans\necho lemon\n",
		}),
		0,
		"line 4: lochNessMonster: command not found")
}

func TestBadCommandInTheMiddle(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"echo tofu\ndate\n",
			"echo beans\necho kale\n",
			"lochNessMonster\n",
			"echo hasta\necho la vista\n",
		}),
		2,
		"line 14: lochNessMonster: command not found")
}

func TestBadCommandAtEnd(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"echo tofu\ndate\n",
			"echo beans\necho kale\n",
			"echo hasta\necho la vista\n",
			"echo hey\nlochNessMonster\n",
		}),
		3,
		"line 20: lochNessMonster: command not found")
}

func TestTimeOutAtStart(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"date\nsleep " + sleep.String() + "\necho kale",
			"echo beans\necho lemon\n",
		}),
		0,
		scanner.MsgTimeout)
}

func TestTimeOutInTheMiddle(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"echo beans\necho lemon\n",
			"date\nsleep " + sleep.String() + "\necho kale",
			"echo tofu\ndate\n",
		}),
		1,
		scanner.MsgTimeout)
}

func TestTimeOutAtTheEnd(t *testing.T) {
	checkFail(
		t,
		doIt([]string{
			"echo beans\necho lemon\n",
			"echo tofu\ndate\n",
			"date\nsleep " + sleep.String() + "\n",
		}),
		2,
		scanner.MsgTimeout)
}
