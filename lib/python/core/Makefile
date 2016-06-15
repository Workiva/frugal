SHELL := /bin/bash
PYTHON := python
PIP := pip
HOME := $(PWD)

all: deps

clean:
	find . -name "*.py[co]" -delete

test: clean unit

deps: 
	$(PIP) install -Ur requirements_dev.txt 

sniffer:
	sniffer -x--with-doctest -x--logging-level=ERROR -x--with-coverage -x--cover-package=frugal

unit:
	nosetests --logging-level=ERROR --with-coverage --cover-package=frugal

xunit:
	nosetests --logging-level=ERROR --with-coverage --cover-package=frugal --with-xunit --xunit-file=unit_tests.xml

install:
	$(PYTHON) setup.py sdist
