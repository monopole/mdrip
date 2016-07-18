package model

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/monopole/mdrip/scanner"
	"github.com/monopole/mdrip/util"
)

// Program is a list of scripts, each from their own file.
type Program struct {
	blockTimeout time.Duration
	label        Label
	scripts      []*script
}

const tmplNameProgram = "program"
const tmplBodyProgram = `
{{define "` + tmplNameProgram + `"}}
<p>
script count = {{.ScriptCount}}
</p>
{{range $i, $s := .Scripts}}
  <div data-id="{{$i}}">
  {{ template "` + tmplNameScript + `" $s }}
{{end}}
<p>
That's it.
</p>
{{end}}
`

var templates = template.Must(
	template.New("main").Parse(
		tmplBodyCommandBlock +
			tmplBodyScript +
			tmplBodyProgram))

func NewProgram(timeout time.Duration, label Label) *Program {
	return &Program{timeout, label, []*script{}}
}

func (p *Program) Add(s *script) *Program {
	p.scripts = append(p.scripts, s)
	return p
}

// Exported only for the template.
func (p *Program) Scripts() []*script {
	return p.scripts
}

func (p *Program) ScriptCount() int {
	return len(p.scripts)
}

// PrintNormal simply prints the contents of a program.
func (p Program) PrintNormal(w io.Writer) {
	for _, s := range p.scripts {
		s.Print(w, p.label, 0)
	}
	fmt.Fprintf(w, "echo \" \"\n")
	fmt.Fprintf(w, "echo \"All done.  No errors.\"\n")
}

// PrintPreambled emits the first n blocks of a script normally, then
// emits the n blocks _again_, as well as all the remaining scripts,
// so that they run in a subshell with signal handling.
//
// This allows the aggregrate script to be structured as 1) a preamble
// initialization script that impacts the environment of the active
// shell, followed by 2) a script that executes as a subshell that
// exits on error.  An exit in (2) won't cause the active shell (most
// likely a terminal) to close.
//
// It's up to the markdown author to assure that the n blocks can
// always complete without exit on error because they will run in the
// existing terminal.  Hence these blocks should just set environment
// variables and/or define shell functions.
//
// The goal is to let the user both modify their existing terminal
// environment, and run remaining code in a trapped subshell, and
// survive any errors in that subshell with a modified environment.
func (p Program) PrintPreambled(w io.Writer, n int) {
	// Write the first n blocks if the first script normally.
	p.scripts[0].Print(w, p.label, n)
	// Followed by everything appearing in a bash subshell.
	hereDocName := "HANDLED_SCRIPT"
	fmt.Fprintf(w, " bash -euo pipefail <<'%s'\n", hereDocName)
	fmt.Fprintf(w, "function handledTrouble() {\n")
	fmt.Fprintf(w, "  echo \" \"\n")
	fmt.Fprintf(w, "  echo \"Unable to continue!\"\n")
	fmt.Fprintf(w, "  exit 1\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "trap handledTrouble INT TERM\n")
	p.PrintNormal(w)
	fmt.Fprintf(w, "%s\n", hereDocName)
}

func write(output string, writer io.Writer) {
	n, err := writer.Write([]byte(output))
	if err != nil {
		glog.Fatalf("Could not write %d bytes: %v", len(output), err)
	} else if n != len(output) {
		glog.Fatalf("Expected to write %d bytes, wrote %d", len(output), n)
	}
}

// check reports the error fatally if it's non-nil.
func check(msg string, err error) {
	if err != nil {
		fmt.Printf("Problem with %s: %v\n", msg, err)
		glog.Fatal(err)
	}
}

// accumulateOutput returns a channel to which it writes objects that
// contain what purport to be the entire output of one command block.
//
// To do so, it accumulates strings off a channel representing command
// block output until the channel closes, or until a string arrives
// that matches a particular pattern.
//
// On the happy path, strings are accumulated and every so often sent
// out with a success == true flag attached.  This continues until the
// input channel closes.
//
// On a sad path, an accumulation of strings is sent with a success ==
// false flag attached, and the function exits early, before it's
// input channel closes.
func accumulateOutput(prefix string, in <-chan string) <-chan *BlockOutput {
	out := make(chan *BlockOutput)
	var accum bytes.Buffer
	go func() {
		defer close(out)
		for line := range in {
			if strings.HasPrefix(line, scanner.MsgTimeout) {
				accum.WriteString("\n" + line + "\n")
				accum.WriteString("A subprocess might still be running.\n")
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Timeout return.", prefix)
				}
				out <- NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgError) {
				accum.WriteString(line + "\n")
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Error return.", prefix)
				}
				out <- NewFailureOutput(accum.String())
				return
			}
			if strings.HasPrefix(line, scanner.MsgHappy) {
				if glog.V(2) {
					glog.Info("accumulateOutput %s: %s", prefix, line)
				}
				out <- NewSuccessOutput(accum.String())
				accum.Reset()
			} else {
				if glog.V(2) {
					glog.Info("accumulateOutput %s: Accumulating [%s]", prefix, line)
				}
				accum.WriteString(line + "\n")
			}
		}

		if glog.V(2) {
			glog.Info("accumulateOutput %s: <--- This channel has closed.", prefix)
		}
		trailing := strings.TrimSpace(accum.String())
		if len(trailing) > 0 {
			if glog.V(2) {
				glog.Info(
					"accumulateOutput %s: Erroneous (missing-happy) output [%s]", prefix, accum.String())
			}
			out <- NewFailureOutput(accum.String())
		} else {
			if glog.V(2) {
				glog.Info("accumulateOutput %s: Nothing trailing.", prefix)
			}
		}
	}()
	return out
}

// userBehavior acts like a command line user.
//
// TODO(monopole): update the comments, as this function no longer writes to stdin.
// See https://github.com/monopole/mdrip/commit/a7be6a6fb62ccf8dfe1c2906515ce3e83d0400d7
//
// It writes command blocks to shell, then waits after  each block to
// see if the block worked.  If the block appeared to complete without
// error, the routine sends the next block, else it exits early.
func (p *Program) userBehavior(stdOut, stdErr io.ReadCloser) (errResult *RunResult) {

	chOut := scanner.BuffScanner(p.blockTimeout, "stdout", stdOut)
	chErr := scanner.BuffScanner(1*time.Minute, "stderr", stdErr)

	chAccOut := accumulateOutput("stdOut", chOut)
	chAccErr := accumulateOutput("stdErr", chErr)

	errResult = NewRunResult()
	for _, script := range p.scripts {
		numBlocks := len(script.Blocks())
		for i, block := range script.Blocks() {
			glog.Info("Running %s (%d/%d) from %s\n",
				block.Name(), i+1, numBlocks, script.FileName())
			if glog.V(2) {
				glog.Info("userBehavior: sending \"%s\"", block.Code())
			}

			result := <-chAccOut

			if result == nil || !result.Succeeded() {
				// A nil result means stdout has closed early because a
				// sub-subprocess failed.
				if result == nil {
					if glog.V(2) {
						glog.Info("userBehavior: stdout Result == nil.")
					}
					// Perhaps chErr <- scanner.MsgError +
					//   " : early termination; stdout has closed."
				} else {
					if glog.V(2) {
						glog.Info("userBehavior: stdout Result: %s", result.Output())
					}
					errResult.setOutput(result.Output()).setMessage(result.Output())
				}
				errResult.setFileName(script.FileName()).setIndex(i).setBlock(block)
				fillErrResult(chAccErr, errResult)
				return
			}
		}
	}
	glog.Info("All done, no errors triggered.\n")
	return
}

// fillErrResult fills an instance of RunResult.
func fillErrResult(chAccErr <-chan *BlockOutput, errResult *RunResult) {
	result := <-chAccErr
	if result == nil {
		if glog.V(2) {
			glog.Info("userBehavior: stderr Result == nil.")
		}
		errResult.setProblem(errors.New("unknown"))
		return
	}
	errResult.setProblem(errors.New(result.Output())).setMessage(result.Output())
	if glog.V(2) {
		glog.Info("userBehavior: stderr Result: %s", result.Output())
	}
}

// RunInSubShell runs command blocks in a subprocess, stopping and
// reporting on any error.  The subprocess runs with the -e flag, so
// it will abort if any sub-subprocess (any command) fails.
//
// Command blocks are strings presumably holding code from some shell
// language.  The strings may be more complex than single commands
// delimitted by linefeeds - e.g. blocks that operate on HERE
// documents, or multi-line commands using line continuation via '\',
// quotes or curly brackets.
//
// This function itself is not a shell interpreter, so it has no idea
// if one line of text from a command block is an individual command
// or part of something else.
//
// Error reporting works by discarding output from command blocks that
// succeeded, and only reporting the contents of stdout and stderr
// when the subprocess exits on error.
func (p *Program) RunInSubShell() (result *RunResult) {
	// Write program to a file to be executed.
	scriptFile, err := ioutil.TempFile("", "mdrip-script-")
	check("create temp file", err)
	check("chmod temp file", os.Chmod(scriptFile.Name(), 0744))
	for _, script := range p.scripts {
		for _, block := range script.Blocks() {
			write(block.Code().String(), scriptFile)
			write("\n", scriptFile)
			write("echo "+scanner.MsgHappy+" "+block.Name().String()+"\n", scriptFile)
		}
	}
	if glog.V(2) {
		glog.Info("RunInSubShell: running commands from %s", scriptFile.Name())
	}
	defer func() {
		check("delete temp file", os.Remove(scriptFile.Name()))
	}()

	// Adding "-e" to force the subshell to die on any error.
	shell := exec.Command("bash", "-e", scriptFile.Name())

	stdIn, err := shell.StdinPipe()
	check("in pipe", err)
	check("close shell's stdin", stdIn.Close())

	stdOut, err := shell.StdoutPipe()
	check("out pipe", err)

	stdErr, err := shell.StderrPipe()
	check("err pipe", err)

	err = shell.Start()
	check("shell start", err)

	pid := shell.Process.Pid
	if glog.V(2) {
		glog.Info("RunInSubShell: pid = %d", pid)
	}
	pgid, err := util.GetProcesssGroupId(pid)
	if err == nil {
		if glog.V(2) {
			glog.Info("RunInSubShell:  pgid = %d", pgid)
		}
	}

	result = p.userBehavior(stdOut, stdErr)

	if glog.V(2) {
		glog.Info("RunInSubShell:  Waiting for shell to end.")
	}
	waitError := shell.Wait()
	if result.Problem() == nil {
		result.setProblem(waitError)
	}
	if glog.V(2) {
		glog.Info("RunInSubShell:  Shell done.")
	}

	// killProcesssGroup(pgid)
	return
}

// Serve offers an http service at the given port.
func (p *Program) Serve(port int) {
	http.HandleFunc("/", p.foo)
	http.HandleFunc("/favicon.ico", p.favicon)
	http.HandleFunc("/image", p.image)
	http.HandleFunc("/runblock", p.runblock)
	http.HandleFunc("/q", p.quit)
	host := "localhost:" + strconv.Itoa(port)
	glog.Info("Serving at " + host)
	glog.Fatal(http.ListenAndServe(host, nil))
}

func (p *Program) favicon(w http.ResponseWriter, r *http.Request) {
	Lissajous(w, 7, 3, 1)
}

func (p *Program) image(w http.ResponseWriter, r *http.Request) {
	Lissajous(w,
		getIntParam("s", r, 300),
		getIntParam("c", r, 30),
		getIntParam("n", r, 100))
}

func getIntParam(n string, r *http.Request, d int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(n))
	if err != nil {
		return d
	}
	return v
}

func (p *Program) quit(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}

const headerHtml = `
<head>
<style type="text/css">
</style>
<script type="text/javascript">
  function getId(el) {
    return el.getAttribute("data-id");
  }
  function setButtonsDisabled(value) {
    var buttons = document.getElementsByTagName('input');
    for (var i = 0; i < buttons.length; i++) {
      buttons[i].disabled = value;
    }
  }
  function showOutput(blockEl, text) {
    var c = blockEl.children;
    for (var i = 0; i < c.length; i++) {
      child = c[i];
      if (child.getAttribute("data-result")) {
        child.innerHTML = "<pre> " + text + " </pre>";
      }
    }
  }
  requestRunning = false
  function onRunBlockClick(event) {
    if (!(event && event.target)) {
      alert("no event!");
      return
    }
    if (requestRunning) {
      alert("busy!");
      return
    }
    requestRunning = true;
    setButtonsDisabled(true)
    var b = event.target;
    blockId = getId(b.parentNode);
    scriptId = getId(b.parentNode.parentNode);
    var oldColor =  b.style.color;
    var oldValue =  b.value;
    b.style.color = 'red';
    b.value = 'running...';
    showOutput(b.parentNode, "running...")
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
      if (xhttp.readyState == XMLHttpRequest.DONE) {
        b.style.color = oldColor;
        b.value = oldValue;
        showOutput(
           b.parentNode,
           (xhttp.status == 200 ?
               xhttp.responseText :
               "status = " + xhttp.status));
        requestRunning = false;
        setButtonsDisabled(false)
      }
    };
    xhttp.open("GET", "/runblock?s=" + scriptId + "&b=" + blockId, true);
    xhttp.send();
  }
</script>
</head>
`

func (p *Program) runblock(w http.ResponseWriter, r *http.Request) {
	indexScript := getIntParam("s", r, -1)
	indexBlock := getIntParam("b", r, -1)
	glog.Info("Run called; s=", indexScript, " b=", indexBlock)
	code := p.scripts[indexScript].Blocks()[indexBlock].Code()
	time.Sleep(3 * time.Second)
	glog.Info("Run done.")
	// TODO(jregan): replace this with result of _running_ the code
	// (not the actual code).
	fmt.Fprint(w, code)
}

func (p *Program) foo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<html>")
	fmt.Fprint(w, headerHtml)
	fmt.Fprintln(w, "<body>")
	if err := templates.ExecuteTemplate(w, tmplNameProgram, p); err != nil {
		glog.Fatal(err)
	}
	fmt.Fprintln(w, "</body>")
	fmt.Fprintln(w, "</html>")
}
