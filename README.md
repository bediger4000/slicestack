## Obtain and fix up `package routine`
```
$ cd $GOPATH/src
$ git clone https://github.com/timandy/routine.git
```

Add the following function to file `routine/g.go`:
```
func Getg() unsafe.Pointer {
    return unsafe.Pointer(getg())
}
```
