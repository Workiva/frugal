THIS_REPO := github.com/Workiva/frugal

all: unit

clean:
	@rm -rf /tmp/frugal

check-cli:
	go vet ./...

unit: clean unit-cli unit-go unit-java unit-py2 unit-py3

unit-cli:
	go test ./... -race

# update the expected generated files; for those times when you change the generation
unit-cli-copy:
	go test ./compiler -copy-files

unit-go:
	$(MAKE) -C $(PWD)/lib/go check-local test-local

unit-java:
	$(MAKE) -C $(PWD)/lib/java check-local test-local

unit-py:
	python -m venv /tmp/frugal && \
	. /tmp/frugal/bin/activate && \
	$(MAKE) -C $(PWD)/lib/python deps-local check-local test-local && \
	deactivate

install:
	go install

.PHONY: \
	all \
	clean \
	unit \
	unit-cli \
	unit-go \
	unit-java \
	unit-py
