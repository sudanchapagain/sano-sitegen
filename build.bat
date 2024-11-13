@echo off

set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

set GOFILE=.\main.go

set OUTPUT_BINARY=sanositegen-linux-amd64

go build -o %OUTPUT_BINARY% %GOFILE%

set GOOS=
set GOARCH=
set CGO_ENABLED=

echo Build complete!
pause
