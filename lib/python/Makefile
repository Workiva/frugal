SHELL := /bin/bash
PYTHON := python
PIP := pip

all: deps

clean:
	find . -name "*.py[co]" -delete

distclean: clean
	rm -rf $(BUILD_DIR)
	rm -rf $(LIBS_DIR)

test: clean unit

deps-py2:
	$(PIP) install -Ur requirements_dev_py2.txt

deps-py3:
	$(PIP) install -Ur requirements_dev_py3.txt

sniffer:
	sniffer -x--with-doctest -x--logging-level=ERROR -x--with-coverage -x--cover-package=frugal

unit-py2:
	nosetests --exclude-dir=frugal/aio --exclude-dir=frugal/tests/aio --logging-level=ERROR --with-coverage --cover-package=frugal

unit-py3:
	nosetests --exclude-dir=frugal/tornado --exclude-dir=frugal/tests/tornado --logging-level=ERROR --with-coverage --cover-package=frugal

xunit-py2:
	nosetests --exclude-dir=frugal/aio --exclude-dir=frugal/tests/aio --logging-level=ERROR --with-coverage --cover-package=frugal --with-xunit --xunit-file=unit_tests_py2.xml

xunit-py3:
	nosetests --exclude-dir=frugal/tornado --exclude-dir=frugal/tests/tornado --logging-level=ERROR --with-coverage --cover-package=frugal --with-xunit --xunit-file=unit_tests_py3.xml

install:
	$(PYTHON) setup.py sdist
