#!/bin/bash

echo "Cleaning up any potentially old generated code"
rm -rf dart/gen-dart
rm -rf go/gen-go
rm -rf java/gen-java
rm -rf python/gen-py.tornado

echo "Generating code in dart, go, java, and python for the frugal files"
frugal -r --gen dart -out='dart/gen-dart' event.frugal
frugal -r --gen go:package_prefix=github.com/Workiva/frugal/example/go/gen-go/ -out='go/gen-go' event.frugal
frugal -r --gen java:generated_annotations=undated -out='java/gen-java' event.frugal
frugal -r --gen py -out='python/gen-py' event.frugal
frugal -r --gen py:tornado -out='python.tornado/gen-py.tornado' event.frugal
frugal -r --gen py:asyncio -out='python.asyncio/gen-py.asyncio' event.frugal

echo "Done!"
