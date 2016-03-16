# Apache Thrift - integration test suite

This is the cross everything integration test suite for Frugal.

## Run

### A. Using Skynet

Tests will be executed on Skynet for any PR.

### B. Using test script directly

Alternatively, you can invoke [test.py](test.py) directly. 

For example, if you changed something in `java` library and need to verify
the patch, you can skip everything except `java` itself and some reference
implementation like this:

    test/integration/test.py --server java


--------------------------------------------------------------------------------
CONTINUE EDITING BELOW THIS POINT

## Test case definition file

The cross test cases are defined in [tests.json](tests.json).
The root element is collection of test target definitions.
Each test target definition looks like this:

    {
      "name": "somelib",

      "client": {
        "command": ["somelib_client_executable"],
        "workdir": "somelib/bin",
        "protocols": ["binary"],
        "transports": ["buffered"],
        "sockets": ["ip"],
      },
      "server": {
        "command": ["somelib_server_executable"],
        "workdir": "somelib/bin",
        "protocols": ["binary"],
        "transports": ["buffered"],
        "sockets": ["ip", "ip-ssl"],
      }
    }

Either client or server definition or both should be present.

Parameters that are common to both `client` and `server` can be put to target
definition root:

    {
      "name": "somelib",

      "workdir": "somelib/bin",
      "protocols": ["binary"],
      "transports": ["buffered"],
      "sockets": ["ip"],

      "client": { "command": ["somelib_client_executable"] },
      "server": {
        "command": ["somelib_server_executable"],
        "sockets": ["ip-ssl"]
      }
    }


## Test executable specification

### Command line parameters

Unit tests for languages are usually located under lib/<lang>/test/
cross language tests according to [ThriftTest.thrift](ThriftTest.thrift) shall be
provided for every language including executables with the following command
line interface:

**Server command line interface:**

    $ ./cpp/TestServer -h
    Allowed options:
      -h [ --help ]               produce help message
      --port arg (=9090)          Port number to listen
      --domain-socket arg         Unix Domain Socket (e.g. /tmp/ThriftTest.thrift)
      --named-pipe arg            Windows Named Pipe (e.g. MyThriftPipe)
      --server-type arg (=simple) type of server, "simple", "thread-pool",
                                  "threaded", or "nonblocking"
      --transport arg (=buffered) transport: buffered, framed, http, anonpipe
      --protocol arg (=binary)    protocol: binary, compact, json
      --ssl                       Encrypted Transport using SSL
      --processor-events          processor-events
      -n [ --workers ] arg (=4)   Number of thread pools workers. Only valid for
                              thread-pool server type

**Client command line interface:**

    $ ./cpp/TestClient -h
    Allowed options:
      -h [ --help ]               produce help message
      --host arg (=localhost)     Host to connect
      --port arg (=9090)          Port number to connect
      --domain-socket arg         Domain Socket (e.g. /tmp/ThriftTest.thrift),
                                  instead of host and port
      --named-pipe arg            Windows Named Pipe (e.g. MyThriftPipe)
      --anon-pipes hRead hWrite   Windows Anonymous Pipes pair (handles)
      --transport arg (=buffered) Transport: buffered, framed, http, evhttp
      --protocol arg (=binary)    Protocol: binary, compact, json
      --ssl                       Encrypted Transport using SSL
      -n [ --testloops ] arg (=1) Number of Tests
      -t [ --threads ] arg (=1)   Number of Test threads

If you have executed the **make check** or **make cross** then you will be able to browse
[gen-html/ThriftTest.html](gen-html/ThriftTest.html) with the test documentation.

### Return code

The return code (exit code) shall be 0 on success, or an integer in the range 1 - 255 on errors.
In order to signal failed tests, the return code shall be composed from these bits to indicate
failing tests:

      #define TEST_BASETYPES     1  // 0000 0001
      #define TEST_STRUCTS       2  // 0000 0010
      #define TEST_CONTAINERS    4  // 0000 0100
      #define TEST_EXCEPTIONS    8  // 0000 1000
      #define TEST_NOTUSED     240  // 1111 0000 (reserved bits)

Tests that have not been executed at all count as errors.

**Example:**

During tests, the test client notices that some of the Struct tests fail.
Furthermore, due to some other problem none of the Exception tests is executed.
Therefore, the test client returns the code `10 = 2 | 8`, indicating the failure
of both test 2 (TEST_STRUCTS) and test 8 (TEST_EXCEPTIONS).


## SSL
Test Keys and Certificates are provided in multiple formats under the following
directory [test/keys](keys)
