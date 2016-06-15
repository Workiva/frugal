SHELL := /bin/bash
PYTHON := python
PIP := pip
HOME := $(PWD)
BUILD_DIR := .build

all: deps

clean:
	find . -name "*.py[co]" -delete

distclean: clean
	rm -rf $(BUILD_DIR)
	rm -rf $(LIBS_DIR)

test: clean unit
deps: py_dev_deps py_deploy_deps

py_deploy_deps: $(BUILD_DIR)/pip-deploy.out

py_dev_deps: $(BUILD_DIR)/pip-dev.out

$(BUILD_DIR)/pip-deploy.out: requirements.txt
	@mkdir -p $(BUILD_DIR)
	$(PIP) install -Ur $< && touch $@

$(BUILD_DIR)/pip-dev.out: requirements_dev.txt
	@mkdir -p $(BUILD_DIR)
	$(PIP) install -Ur $< && touch $@

sniffer:
	sniffer -x--with-doctest -x--logging-level=ERROR -x--with-coverage -x--cover-package=frugal

unit:
	nosetests --logging-level=ERROR --with-coverage --cover-package=frugal

xunit:
	nosetests --logging-level=ERROR --with-coverage --cover-package=frugal --with-xunit --xunit-file=unit_tests.xml

install:
	$(PYTHON) setup.py sdist

test_local:
	./run_tests_local.sh
