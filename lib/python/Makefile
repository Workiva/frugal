SHELL := /bin/bash
PYTHON := python
PIP := pip
EXCLUDE_GAE := --exclude-dir=frugal/gae --exclude-dir=frugal/tests/gae
EXCLUDE_TORNADO := --exclude-dir=frugal/tornado --exclude-dir=frugal/tests/tornado
EXCLUDE_AIO := --exclude-dir=frugal/aio --exclude-dir=frugal/tests/aio
COVER := --with-coverage --cover-package=frugal-py3
LEVEL := --logging-level=ERROR

all: deps

deps: deps-tornado deps-gae

clean:
	find . -name "*.py[co]" -delete

distclean: clean
	rm -rf $(BUILD_DIR)
	rm -rf $(LIBS_DIR)

deps-py3:
	$(PIP) install -Ur requirements_python_3.txt

deps-tornado:
	$(PIP) install -Ur requirements_dev_tornado.txt

deps-gae:
	$(PIP) install -Ur requirements_dev_gae.txt

deps-asyncio:
	$(PIP) install -Ur requirements_dev_asyncio.txt

sniffer:
	sniffer -x--with-doctest -x--logging-level=ERROR -x--with-coverage -x--cover-package=frugal

flake8-py3:
	flake8 --exclude=tests,frugal/tornado,frugal/coverage,build

unit-py3:
	nosetests $(EXCLUDE_TORNADO) $(EXCLUDE_GAE) $(EXCLUDE_AIO) $(LEVEL) $(COVER)

xunit-py3:
	nosetests $(EXCLUDE_TORNADO) $(EXCLUDE_GAE) $(LEVEL) $(COVER) --with-xunit --xunit-file=unit_tests_py3.xml

install:
	$(PYTHON) setup.py sdist

all-local: clean-local deps-local check-local test-local

clean-local:
	find . -name "*.py[co]" -delete

deps-local:
	$(PIP) install -Ur requirements_python_3.txt
	$(PIP) install -Ur requirements_dev_tornado.txt
	$(PIP) install -Ur requirements_dev_gae.txt
	$(PIP) install -Ur requirements_dev_asyncio.txt

check-local:
	flake8 --exclude=tests,frugal/tornado

test-local:
	nosetests $(EXCLUDE_TORNADO) $(EXCLUDE_GAE) $(EXCLUDE_AIO) $(LEVEL)
