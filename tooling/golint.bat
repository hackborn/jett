REM install
go get -u github.com/kisielk/errcheck
go get github.com/jgautheron/goconst/cmd/goconst
go get -u golang.org/x/tools/...

REM run
goconst.exe ./../...
errcheck.exe ./../...
REM gocyclo.exe ./../...
gotype.exe ./..