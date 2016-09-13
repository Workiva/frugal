#!/usr/bin/env bash

# Python Dependencies
cd $GOPATH/src/github.com/Workiva/frugal/lib/python
pip install -e ".[tornado]"
pip install -r requirements_dev_tornado.txt
