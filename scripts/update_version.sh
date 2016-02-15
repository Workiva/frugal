#!/usr/bin/env bash

# WARNING: ONLY RUN THIS FROM ROOT IN THE FRUGAL DIRECTORY!!
# WARNING: COMMIT AND PUSH ANY LOCAL CHANGES TO BE SAFE!!
# THIS IS WRITTEN FOR OSX, IF ON LINUX REMOVE THE '' AFTER -i IN THE 
# REPLACE FUNCTION
# YOU WILL LIKELY NEED TO ADD
#   export LC_CTYPE=C 
#   export LANG=C
# TO YOUR BASH PROFILE

CURRENT=$1
NEXT=$2
ROOT=$PWD

function replace {
    find . -type f -exec sed -i '' 's/'$CURRENT'/'$NEXT'/g' {} +
}
    
# Change versions in the compiler
cd $ROOT/compiler
replace

# Change versions in the lib
cd $ROOT/lib
replace

# Change versions in the test
cd $ROOT/test
replace

# Change versions in the example
cd $ROOT/example
replace

# Return home, your work is done!
cd $ROOT
