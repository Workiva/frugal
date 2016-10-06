PROJECT_ROOT = ../..
IDL_DIR = $(PROJECT_ROOT)/examples
ifneq ("$(wildcard $(PROJECT_ROOT)/frugal)","")
FRUGAL_PATH = $(PROJECT_ROOT)/frugal
else
FRUGAL_PATH = frugal
endif

generate-java:
	@-rm -rf target
	find $(IDL_DIR) -type f -name '*.frugal' \
		-exec $(FRUGAL_PATH) -out target/generated-test-sources/generated/frugal/ --gen java {} \;
