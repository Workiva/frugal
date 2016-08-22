IDL_DIR = idl/
GO = go/gen-go
PYTHON = python/gen-py
PYTHON_TORNADO = python/gen-py.tornado
DART = dart/gen-dart
JAVA = java/src/main/java/
THIS_REPO = github.com/Workiva/linking-frugal

clean:
	@-rm -rf $(GO)
	@-rm -rf $(PYTHON)
	@-rm -rf $(DART)
	@-rm -rf $(JAVA)

generate: clean
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(GO) --gen go:package_prefix=$(THIS_REPO)/$(GO)/ {} \;
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(JAVA) --gen java {} \;
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(PYTHON) --gen py {} \;
	find $(IDL_DIR) -type f -name '*.frugal' -exec frugal -out $(PYTHON_TORNADO) --gen py:tornado {} \;

