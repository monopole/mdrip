<!-- @bad @good @sleep -->
```
nc -l 8000 &
PID=$!
```

<!-- @good -->
```
kill $PID
```

<!-- @bad -->
```
echo "Don't forget to: killall nc"
```

<!-- @bad @good -->
```
echo "About to trigger a failure."
echo ${DONT_EXIST?} > /dev/null
```

