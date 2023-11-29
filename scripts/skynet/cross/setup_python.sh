#!/usr/bin/env bash

set -ex

cd $GOPATH/src/github.com/Workiva/frugal/lib/python

python2 -m pip install -e ".[tornado]"

python3 -m pip install Cython==0.27.3
python3 -m pip install -e ".[asyncio]"
