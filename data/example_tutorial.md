Follow these instructions to write a Go program.

First create a directory to work in.

<!-- @init @lesson1 @cleanup -->
```
DEMO_DIR=$(mktemp -d)
mkdir -p $DEMO_DIR/src/example
```

Then write a *Go* function:

<!-- @makeAdder @lesson1 -->
```
 cat - <<EOF >$DEMO_DIR/src/example/add.go
package main

func add(x, y int) (int) { return x + y }
EOF
echo "the next command intended to fail"
badCommandToTriggerTestFailure
```

Then write a main program to call it:

<!-- @makeMain @lesson1 -->
```
 cat - <<EOF >$DEMO_DIR/src/example/main.go
package main

import "fmt"

func main() {
    comment this line to avoid compiler error
    fmt.Printf("Calling add on 1 and 2 yields %d.\n", add(1, 2))
}
EOF
echo "The following compile should fail."
GOPATH=$DEMO_DIR go install example
$DEMO_DIR/bin/example
```

Copy/paste the above into a shell to build and run your *Go* program.

Clean up with this command:

<!-- @cleanup @lesson1 @sleep -->
```
/bin/rm -rf $DEMO_DIR
```
