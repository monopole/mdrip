package internal

import (
	"fmt"
	"log/slog"
	"os/exec"
	"time"
)

const (
	indent = "  "
	doing  = "  [x] "
	faking = "  [ ] "
)

type safetyLevel int

const (
	// NoHarmDone signals a command that is harmless to run, so it should
	// always be run. E.g. some read-only command that runs quickly.
	NoHarmDone safetyLevel = iota
	// UndoIsHard signals a command that likely cannot be undone, e.g.
	// a POST to a website.
	UndoIsHard
)

type verbosity int

const (
	Low verbosity = iota
	High
)

type behavior int

const (
	DoIt behavior = iota
	FakeIt
)

// MyRunner runs some program with different arguments, timing the run,
// quiting if a time limit is exceeded.
type MyRunner struct {
	// program is the name of the program to run
	program string
	// env hold name=value strings
	env []string
	// duration is the time limit on a command run.
	duration time.Duration
	// From which directory do we run the commands.
	workDir string
	// Run commands, or merely print commands.
	doIt behavior
	// vb controls whether commands are echoed to stdout
	vb verbosity
	// out hold the output of the most recently run command (stdout and stderr).
	out []byte
}

func NewMyRunner(program string, wd string, doIt behavior, d time.Duration) *MyRunner {
	theProgram, err := exec.LookPath(program)
	if err != nil {
		panic(err)
	}
	return &MyRunner{
		program:  theProgram,
		duration: d,
		workDir:  wd,
		doIt:     doIt,
		vb:       High,
	}
}

func (rn *MyRunner) Out() string {
	return string(rn.out)
}

func (rn *MyRunner) comment(f string) {
	if rn.vb == Low {
		return
	}
	slog.Info(indent + f)
}

func (rn *MyRunner) doing(s string) {
	if rn.vb == Low {
		return
	}
	slog.Info(indent + doing + s)
}

func (rn *MyRunner) faking(s string) {
	if rn.vb == Low {
		return
	}
	slog.Info(indent + faking + s)
}

func (rn *MyRunner) setEnv(m map[string]string) {
	result := make([]string, len(m))
	i := 0
	for k, v := range m {
		result[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	rn.env = result
}

func (rn *MyRunner) run(sl safetyLevel, args ...string) error {
	c := exec.Command(rn.program, args...)
	if rn.doIt == FakeIt && sl != NoHarmDone /* if no harm done, then do it */ {
		rn.faking(c.String())
		rn.out = nil
		return nil
	}
	rn.doing(c.String())
	c.Dir = rn.workDir
	c.Env = rn.env
	return TimedCall(
		c.String(),
		rn.duration,
		func() error {
			var err error
			rn.out, err = c.CombinedOutput()
			if err != nil {
				slog.Error(string(rn.out))
				return fmt.Errorf("failed to run %q: %w", c.String(), err)
			}
			return nil
		})
}
