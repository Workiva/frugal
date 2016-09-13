#!/usr/bin/env bash
set -e


# Python
virtualenv -p /usr/bin/python /tmp/frugal
source /tmp/frugal/bin/activate
pip install -U pip
cd $FRUGAL_HOME/lib/python
make deps-tornado
make deps-gae
make xunit-py2
deactivate

virtualenv -p /usr/bin/python3.5 /tmp/frugal-py3
source /tmp/frugal-py3/bin/activate
pip install -U pip
cd $FRUGAL_HOME/lib/python
make deps-asyncio
make xunit-py3
make install
mv dist/frugal-*.tar.gz $SMITHY_ROOT
deactivate