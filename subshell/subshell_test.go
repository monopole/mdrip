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
	return model.NewLessonTutForTests("iamafilename", b)
}

func emptyTutorial() model.Tutorial {
	return newTutorial([]*model.BlockTut{})
}

func TestRunnerWithNothing(t *testing.T) {
	if NewSubshell(
		timeout,
		program.NewProgramFromTutorial(
			base.WildCardLabel, emptyTutorial())).Run().Problem() != nil {
		t.Fail()
	}
}

func doIt(lines []string) *RunResult {
	var pgm []*program.BlockPgm
	for _, l := range lines {
		pgm = append(pgm, makeBlock(l))
	}
	lesson := program.NewLessonPgm(base.FilePath("foo"), pgm)
	p := program.NewProgram([]*program.LessonPgm{lesson})
	return NewSubshell(timeout, p).Run()
}

func TestRunnerWithGoodStuff(t *testing.T) {
	result := doIt([]string{
		"echo kale\ndate\n",
		"echo beans\necho apple\n",
		"echo hasta\necho la vista\n",
	})
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
	checkFail(
		t,
		doIt([]string{
			"notagoodcommand\ndate\n",
			"echo beans\necho cheese\n",
		}),
		NoCommandsRunResult(
			NewFailureOutput("dunno"),
			"fileNameTestStartWithABadCommand",
			0,
			"line 1: notagoodcommand: command not found"))
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
		NoCommandsRunResult(
			NewFailureOutput("dunno"),
			"fileNameTestBadCommandInTheMiddle",
			2,
			"line 11: lochNessMonster: command not found"))
}

func TestTimeOut(t *testing.T) {
	// Insert this sleep in a command block.
	// Arrange to sleep for two seconds longer than the timeout.
	sleep := timeout + (2 * time.Second)
	checkFail(
		t,
		doIt([]string{
			"date\nsleep " + sleep.String() + "\necho kale",
			"echo beans\necho cheese\n",
		}),
		NoCommandsRunResult(
			NewFailureOutput("dunno"),
			"fileNameTestTimeOut",
			0,
			scanner.MsgTimeout))
}
