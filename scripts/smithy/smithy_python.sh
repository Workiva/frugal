#!/usr/bin/env bash
set -e

# Python
virtualenv -p /usr/bin/python /tmp/frugal
source /tmp/frugal/bin/activate
pip install -U pip setuptools==39.0.1
cd $FRUGAL_HOME/lib/python
make deps-tornado
make deps-gae
make xunit-py2

# Write dependencies out so that RM is able to track them
# The name of this file is hard coded into Rosie and RM console
pip freeze > $FRUGAL_HOME/python2_pip_deps.txt
make flake8-py2
deactivate

virtualenv -p /usr/bin/python3 /tmp/frugal-py3
source /tmp/frugal-py3/bin/activate
pip3 install -U pip3 setuptools==39.0.1
cd $FRUGAL_HOME/lib/python
make deps-asyncio
make xunit-py3
make flake8-py3
make install
# Write dependencies out so that RM is able to track them
# The name of this file is hard coded into Rosie and RM console
pip3 freeze > $FRUGAL_HOME/python3_pip_deps.txt
mv dist/frugal-*.tar.gz $FRUGAL_HOME

# get coverage report in correct format
coverage xml
mv $FRUGAL_HOME/lib/python/coverage.xml $FRUGAL_HOME/lib/python/coverage_py3.xml

deactivate
