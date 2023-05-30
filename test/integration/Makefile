clean:
	rm -rf test/integration/go/gen/*
	rm -rf test/integration/java/frugal-integration-test/gen-java/*
	rm -rf test/integration/python/tornado/gen_py_tornado/*
	rm -rf test/integration/python/ascynio/gen_py_asyncio/*
	rm -rf test/integration/python/vanilla/gen_py/*
	rm -rf test/integration/dart/gen-dart/*

generate:
	frugal --gen go:package_prefix=github.com/Workiva/frugal/test/integration/go/gen/ -r --out='test/integration/go/gen' frugalTest.frugal
	frugal --gen java -r --out='test/integration/java/frugal-integration-test/gen-java' frugalTest.frugal
	frugal --gen py:tornado -r --out='test/integration/python/tornado/gen_py_tornado' frugalTest.frugal
	frugal --gen py:asyncio -r --out='test/integration/python/aio/gen_py_asyncio' frugalTest.frugal
	frugal --gen py -r -out='test/integration/python/tornado/gen-py' frugalTest.frugal
	frugal --gen dart:use_enums=true -r --out='test/integration/dart/gen-dart' frugalTest.frugal

cross: clean generate
	@:  # TODO: execute cross tests
