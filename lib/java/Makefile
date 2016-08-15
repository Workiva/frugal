IDL_DIR = ../../example/

generate-java:
	@-rm -rf target
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out target/generated-test-sources/frugal/ --gen java {} \;
