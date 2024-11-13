#!/bin/sh

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sanositegen-linux-amd64 ./main.go

if [ $? -eq 0 ]; then
    echo "Build complete!"
else
    echo "Build failed!"
    exit 1
fi
