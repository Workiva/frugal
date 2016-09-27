#!/usr/bin/env bash

# Tornado Dependencies
cd $GOPATH/src/github.com/Workiva/frugal/lib/python
pip install -e ".[tornado]"
pip install -r requirements_dev_tornado.txt

# Asyncio
python3.5 /usr/bin/pip3 install -e ".[asyncio]"
python3.5 /usr/bin/pip3 install -q -r requirements_dev_asyncio.txt