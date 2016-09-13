# Frugal - Cross language test suite

This is the cross language test suite for Frugal.

## Run

Tests will be executed on Skynet for any PR or can be run locally with Skynet-cli (see
https://github.com/workiva/skynet-cli for setup instructions).


### Known Issues

* 	Binary calls across json protocol are serialized differently between go and java.
    https://github.com/Workiva/frugal/issues/412