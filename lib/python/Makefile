SHELL := /bin/bash
PYTHON := python
PIP := pip
PIP3 := pip3
EXCLUDE_GAE := --exclude-dir=frugal/gae --exclude-dir=frugal/tests/gae
EXCLUDE_TORNADO := --exclude-dir=frugal/tornado --exclude-dir=frugal/tests/tornado
EXCLUDE_AIO := --exclude-dir=frugal/aio --exclude-dir=frugal/tests/aio
COVER := --with-coverage --cover-package=frugal-py3
COVER_PY2 := --with-coverage --cover-package=frugal
LEVEL := --logging-level=ERROR

all: deps

clean:
	find . -name "*.py[co]" -delete

deps: deps-tornado deps-gae

distclean: clean
	rm -rf $(BUILD_DIR)
	rm -rf $(LIBS_DIR)

deps-py2:
	$(PIP) install -Ur ../python2/requirements_python_2.txt

deps-py3:
	$(PIP3) install -Ur requirements_python_3.txt
	
deps-tornado:
	$(PIP) install -Ur requirements_dev_tornado.txt

deps-gae:
	$(PIP) install -Ur requirements_dev_gae.txt

deps-asyncio:
	$(PIP3) install -Ur requirements_dev_asyncio.txt

sniffer:
	sniffer -x--with-doctest -x--logging-level=ERROR -x--with-coverage -x--cover-package=frugal

flake8-py2:
	flake8 --exclude=tests,frugal/aio,build

flake8-py3:
	flake8 --exclude=tests,frugal/tornado,frugal/coverage,build

unit-py2:
	nosetests $(EXCLUDE_AIO) $(LEVEL) $(COVER_PY2)

unit-py3:
	nosetests $(EXCLUDE_TORNADO) $(EXCLUDE_GAE) $(LEVEL) $(COVER)

xunit-py2:
	nosetests $(EXCLUDE_AIO) $(LEVEL) $(COVER_PY2) --with-xunit --xunit-file=unit_tests_py2.xml

xunit-py3:
	nosetests $(EXCLUDE_TORNADO) $(EXCLUDE_GAE) $(LEVEL) $(COVER) --with-xunit --xunit-file=unit_tests_py3.xml

xunit-std-py2:
	nosetests $(EXCLUDE_GAE) $(EXCLUDE_TORNADO) $(EXCLUDE_AIO) $(LEVEL) $(COVER_PY2) --with-xunit --xunit-file=unit_tests_std_py2.xml

install:
	$(PYTHON) setup.py sdist
