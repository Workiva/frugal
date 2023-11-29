#!/usr/bin/env bash
set -e

# Python
virtualenv -p /usr/bin/python /tmp/frugal
source /tmp/frugal/bin/activate
pip install -U pip setuptools==39.0.1

pushd lib/python
make deps-tornado
#all dependent packages that are tied to the python2 environment can be refernced here
make deps-py2

make deps-gae
make xunit-py2

# Write dependencies out so that RM is able to track them
# The name of this file is hard coded into Rosie and RM console
pip freeze > python2_pip_deps.txt
make flake8-py2
deactivate
popd

virtualenv -p /usr/bin/python3 /tmp/frugal-py3
source /tmp/frugal-py3/bin/activate
pip install -U pip setuptools==39.0.1 importlib-metadata==4.13.0

pushd lib/python
#all dependent packages that are seperate from python2 and python3... 
#once move to only python3 then these dependencies can be just put in requirements.txt
make deps-py3
make deps-asyncio
make xunit-py3
make flake8-py3
make install
# Write dependencies out so that RM is able to track them
# The name of this file is hard coded into Rosie and RM console
pip freeze > python3_pip_deps.txt
mv dist/frugal-*.tar.gz ../../

# get coverage report in correct format
coverage xml
popd
mv lib/python/coverage.xml lib/python/coverage_py3.xml

deactivate
