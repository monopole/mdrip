package internal

import (
	"fmt"
	"time"
)

// See https://stackoverflow.com/questions/72307412/applying-errors-is-and-errors-as-on-custom-made-struct-errors
type errTimeOut struct {
	duration time.Duration
	cmd      string
}

func NewErrTimeOut(d time.Duration, c string) *errTimeOut {
	return &errTimeOut{duration: d, cmd: c}
}

func (e *errTimeOut) Error() string {
	return fmt.Sprintf("hit %s timeout running '%s'", e.duration, e.cmd)
}
