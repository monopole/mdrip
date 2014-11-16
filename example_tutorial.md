First do some setup:

<!-- @1 @setup -->
```
export GOPATH=/tmp/play/go
```

Write a *Go* function...

<!-- @1 -->
```
mkdir -p $GOPATH/src/example
 cat - <<EOF >$GOPATH/src/example/add.go
package main

func add(x, y int) (int) { return x + y }
EOF
echo "the next command meant to fail"
badCommandToTriggerTestFailure - remove to make it work
```

...and a main program to call it:

<!-- @1 -->
```
 cat - <<EOF >$GOPATH/src/example/main.go
package main

import "fmt"

func main() {
    fmt.Printf("Calling add on 1 and 2 yields %d.\n", add(1, 2))
}
EOF
go install example
$GOPATH/bin/example
```

Copy/paste the above into a shell to build and run your *Go* program.
