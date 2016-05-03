#!/bin/bash

# Clean up any potentially old generated code
rm -rf dart/gen-dart
rm -rf go/gen-go
rm -rf java/gen-java

# Generate code in each language for the frugal files
frugal -r --gen dart -out='dart/gen-dart' event.frugal
frugal -r --gen go:package_prefix=github.com/Workiva/frugal/example/go/gen-go/ -out='go/gen-go' event.frugal
frugal -r --gen java -out='java/gen-java' event.frugal
