Follow these instructions to write a `Go` program.

Create a directory to work in:

<!-- @createWorkDir @lesson1 -->
```
workDir=$(mktemp -d --tmpdir mdrip_example_XXXXX)
pushd $workDir
```

Make a file with an `add` function:

<!-- @makeAdder @lesson1 -->
```
cat - <<EOF >add.go
package main

func add(x, y int) (int) { return x + y }
EOF
echo "the next command intended to fail"
badCommandToTriggerTestFailure
```

Write a _main program_ to call it:

<!-- @makeMain @lesson1 -->
```
cat - <<EOF >main.go
package main

import "fmt"

func main() {
  comment this line to avoid compiler error
  fmt.Printf("Calling add on 1 and 2 yields %d.\n", add(1, 2))
}
EOF
```

<!-- @defineGoMod @lesson1 -->
```
go mod init myAdder
go mod tidy
```

<!-- @compileMain @lesson1 -->
```
echo "The following compile should fail."
go build .
```

Now you can run main:
<!-- @runMain @lesson1 -->
```
./myAdder
```

Copy/paste the above into a shell to build and run your program.

Clean up with this command:

<!-- @cleanup @lesson1 -->
```
popd
/bin/rm -rf ${workDir}
```
