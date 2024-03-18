#!/usr/bin/env bash

set -ex

if [ -z "${IN_SKYNET_CLI+yes}" ]; then
    mkdir /python
    tar -xzf ${SKYNET_APPLICATION_FRUGAL_PYPI} -C /python
    cd /python/frugal*
else
    cd $GOPATH/src/github.com/Workiva/frugal/lib/python
fi

#python2 -m pip install -e ".[tornado]"

python -m pip install Cython==0.27.3
python -m pip install -e ".[asyncio]"
