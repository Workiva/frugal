THIS_REPO := github.com/Workiva/frugal

all: unit

clean:
	@rm -rf /tmp/frugal
	@rm -rf /tmp/frugal-py3

unit: clean unit-cli unit-go unit-java unit-py2 unit-py3

unit-cli:
	go test ./test -race

unit-go:
	cd lib/go && glide install && go test -v -race 

unit-java:
	mvn -f lib/java/pom.xml checkstyle:check clean verify

unit-py2:
	virtualenv -p /usr/bin/python /tmp/frugal && \
	. /tmp/frugal/bin/activate && \
	$(MAKE) -C $(PWD)/lib/python deps-tornado deps-gae xunit-py2 flake8-py2 &&\
	deactivate

unit-py3:
	virtualenv -p python3 /tmp/frugal-py3 && \
	. /tmp/frugal-py3/bin/activate && \
	$(MAKE) -C $(PWD)/lib/python deps-asyncio xunit-py3 flake8-py3 && \
	deactivate

.PHONY: \
	all \
	clean \
	unit \
	unit-cli \
	unit-go \
	unit-java \
	venv-py2 \
	venv-py3 \
	unit-py2 \
	unit-py3
