SHELL := /bin/bash
PYTHON := python
PIP := pip

all: deps

clean:
	find . -name "*.py[co]" -delete

deps: deps-tornado deps-gae

distclean: clean
	rm -rf $(BUILD_DIR)
	rm -rf $(LIBS_DIR)

deps-tornado:
	$(PIP) install -Ur requirements_dev_tornado.txt

deps-gae:
	$(PIP) install -Ur requirements_dev_gae.txt

deps-asyncio:
	$(PIP) install -Ur requirements_dev_asyncio.txt

sniffer:
	sniffer -x--with-doctest -x--logging-level=ERROR -x--with-coverage -x--cover-package=frugal

flake8-py2:
	flake8 --exclude=tests,frugal/aio

flake8-py3:
	flake8 --exclude=tests,frugal/tornado

unit-py2:
	nosetests --exclude-dir=frugal/aio --exclude-dir=frugal/tests/aio --logging-level=ERROR --with-coverage --cover-package=frugal

unit-py3:
	nosetests --exclude-dir=frugal/tornado --exclude-dir=frugal/tests/tornado --exclude-dir=frugal/gae --exclude-dir=frugal/tests/gae --logging-level=ERROR --with-coverage --cover-package=frugal

xunit-py2:
	nosetests --exclude-dir=frugal/aio --exclude-dir=frugal/tests/aio --logging-level=ERROR --with-coverage --cover-package=frugal --with-xunit --xunit-file=unit_tests_py2.xml

xunit-py3:
	nosetests --exclude-dir=frugal/tornado --exclude-dir=frugal/tests/tornado --exclude-dir=frugal/gae --exclude-dir=frugal/tests/gae --logging-level=ERROR --with-coverage --cover-package=frugal --with-xunit --xunit-file=unit_tests_py3.xml

install:
	$(PYTHON) setup.py sdist
