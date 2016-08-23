THIS_REPO = github.com/Workiva/frugal
IDL_DIR = idl/
GO = go/gen-go
PYTHON_TORNADO = python.tornado/gen-py.tornado
PYTHON_ASYNCIO = python.asyncio/gen-py.asyncio
DART = dart/gen-dart
JAVA = java/src/main/java/

clean:
	@-rm -rf $(GO)
	@-rm -rf $(JAVA)/music/
	@-rm -rf $(PYTHON_TORNADO)
	@-rm -rf $(PYTHON_ASYNCIO)
	@-rm -rf $(DART)

generate: clean
	find $(IDL_DIR) -type f -name '*.frugal' \
		-exec frugal -out $(GO) --gen go:package_prefix=$(THIS_REPO)/$(GO)/ {} \; \
		-exec frugal -out $(JAVA) --gen java {} \; \
		-exec frugal -out $(PYTHON_TORNADO) --gen py:tornado {} \; \
		-exec frugal -out $(PYTHON_ASYNCIO) --gen py:asyncio {} \; \
		-exec frugal -out $(DART) --gen dart {} \;
