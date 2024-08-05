Follow these instructions to write a `Go` program.

Create a directory to work in:

<!-- @createWorkDir -->
```
workDir=$(mktemp -d --tmpdir mdrip_example_XXXXX)
pushd $workDir
```

Make a file with an `add` function:

<!-- @makeAdder -->
```
cat - <<EOF >add.go
package main

func add(x, y int) (int) { return x + y }
EOF
```

Write a _main program_ to call it:

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

<!-- @defineGoMod @goCommand -->
```
go mod init myAdder
go mod tidy
echo Dependencies defined.
badecho Enter go build
```

<!-- @compileMain @goCommand -->
```
go build .
```

Now you can run main:
<!-- @runMain -->
```
./myAdder
```

Copy/paste the above into a shell to build and run your program.

Clean up with this command:

```
popd
/bin/rm -rf ${workDir}
```

<!-- @setEnv -->
```
greeting="Hello world!"
```
