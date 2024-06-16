Follow these instructions to write a `Go` program.

Create a directory to work in:

<!-- @init @lesson1 -->
```
workDir=$(mktemp -d)
mkdir -p ${workDir}/src/example
```

Make a file with an `add` function:

<!-- @makeAdder @lesson1 -->
```
cat - <<EOF >${workDir}/src/example/add.go
package main

func add(x, y int) (int) { return x + y }
EOF
echo "the next command intended to fail"
badCommandToTriggerTestFailure
```

Write a _main program_ to call it:

<!-- @makeMain @lesson1 -->
```
cat - <<EOF >${workDir}/src/example/main.go
package main

import "fmt"

func main() {
  comment this line to avoid compiler error
  fmt.Printf("Calling add on 1 and 2 yields %d.\n", add(1, 2))
}
EOF
echo "The following compile should fail."
GOPATH=${workDir} go install example
${workDir}/bin/example
```

Copy/paste the above into a shell to build and run your program.

Clean up with this command:

<!-- @cleanup @lesson1 -->
```
/bin/rm -rf ${workDir}
```
