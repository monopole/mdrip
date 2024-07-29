package internal

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

const (
	indent = "  "
	doing  = "  [x] "
	faking = "  [ ] "
)

type safetyLevel int

const (
	// Commands that don't hurt, e.g. checking out an existing branch.
	noHarmDone safetyLevel = iota
	// Commands that write, and could be hard to undo.
	undoPainful
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

// runner runs some program with different arguments, timing the run.
type runner struct {
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

func newRunner(program string, wd string, doIt behavior, d time.Duration) *runner {
	theProgram, err := exec.LookPath(program)
	if err != nil {
		panic(err)
	}

	return &runner{
		program:  theProgram,
		duration: d,
		workDir:  wd,
		doIt:     doIt,
		vb:       High,
	}
}

func (rn *runner) Out() string {
	return string(rn.out)
}

func (rn *runner) comment(f string) {
	if rn.vb == Low {
		return
	}
	fmt.Print(indent)
	fmt.Println(f)
}

func (rn *runner) doing(s string) {
	if rn.vb == Low {
		return
	}
	fmt.Print(indent)
	fmt.Print(doing)
	fmt.Println(s)
}

func (rn *runner) faking(s string) {
	if rn.vb == Low {
		return
	}
	fmt.Print(indent)
	fmt.Print(faking)
	fmt.Println(s)
}

func (rn *runner) setEnv(m map[string]string) {
	result := make([]string, len(m))
	i := 0
	for k, v := range m {
		result[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	rn.env = result
}

func (rn *runner) run(sl safetyLevel, args ...string) error {
	c := exec.Command(rn.program, args...)
	if sl != noHarmDone && rn.doIt == FakeIt {
		rn.faking(c.String())
		rn.out = nil
		return nil
	}
	rn.doing(c.String())
	c.Dir = rn.workDir
	rn.comment("workdir = " + c.Dir)
	c.Env = rn.env
	rn.comment("    env = " + strings.Join(c.Env, " "))
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
