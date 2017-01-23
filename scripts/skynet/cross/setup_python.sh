#!/usr/bin/env bash

set -ex

if [ -z "${IN_SKYNET_CLI+yes}" ]; then
    mkdir /python
    tar -xzf ${SKYNET_APPLICATION_FRUGAL_PYPI} -C /python
    cd /python/frugal*
else
    cd $GOPATH/src/github.com/Workiva/frugal/lib/python
fi

pip install -e ".[tornado]"
python3.5 /usr/bin/pip3 install -e ".[asyncio]"

# temporarily install gevent using the requirements until nats is a consumable package
pip install -r /testing/lib/python/requirements_dev_gevent.txt
#pip install -e ".[gevent]"