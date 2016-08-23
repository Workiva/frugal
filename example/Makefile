THIS_REPO = github.com/Workiva/frugal
IDL_DIR = idl/
GO = go/gen-go
PYTHON_TORNADO = python.tornado/gen-py.tornado
DART = dart/gen-dart
JAVA = java/src/main/java/

clean:
	@-rm -rf $(GO)
	@-rm -rf $(JAVA)
	@-rm -rf $(PYTHON_TORNADO)
	@-rm -rf $(DART)

generate: clean
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(GO) --gen go:package_prefix=$(THIS_REPO)/$(GO)/ {} \;
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(JAVA) --gen java {} \;
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(PYTHON_TORNADO) --gen py:tornado {} \;
