THIS_REPO := github.com/Workiva/frugal

all: unit

clean:
	@rm -rf /tmp/frugal
	@rm -rf /tmp/frugal-py3

unit: clean unit-cli unit-go unit-java unit-py2 unit-py3

unit-cli:
	go test ./... -race

# update the expected generated files; for those times when you change the generation
unit-cli-copy:
	go test ./compiler -copy-files

unit-go:
	cd lib/go && go test -v -race

unit-java:
	mvn -f lib/java/pom.xml clean verify

unit-py2:
	python2 -m virtualenv /tmp/frugal && \
	. /tmp/frugal/bin/activate && \
	$(MAKE) -C $(PWD)/lib/python deps-py2 deps-tornado deps-gae xunit-py2 flake8-py2 &&\
	deactivate

unit-py3:
	python -m venv /tmp/frugal-py3 && \
	. /tmp/frugal-py3/bin/activate && \
	$(MAKE) -C $(PWD)/lib/python deps-py3 deps-asyncio xunit-py3 flake8-py3 && \
	deactivate

.PHONY: \
	all \
	clean \
	unit \
	unit-cli \
	unit-go \
	unit-java \
	unit-py2 \
	unit-py3
