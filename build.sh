#!/bin/bash

VERSION=$(git describe --tags --always --dirty --abbrev=7)

echo "Building gosh with version: $VERSION"

go build -ldflags "-X main.Version=$VERSION" -o gosh

if [ $? -eq 0 ]; then
    echo "Build successful! Executable: ./gosh"
else
    echo "Build failed!"
fi