# Frugal - Cross language test suite

This is the cross language test suite for Frugal.

## Run

### A. Using Skynet

Tests will be executed on Skynet for any PR. 

### B. Using test script directly

Alternatively, you can easily run locally by executing ./scripts/run_cross_local.sh

Modify line #34 (python test/integration/test.py --retry-count=0) if you would like to run a specific subset of tests.

</br>

### Known Issues

* Java client to Go server: 2 bytes are added to "testBinary" when using json protocol.