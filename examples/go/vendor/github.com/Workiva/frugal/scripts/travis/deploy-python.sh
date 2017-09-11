#!/usr/bin/env bash
if [ "$TRAVIS_BRANCH" = 'master' ] && [ "$TRAVIS_PULL_REQUEST" == 'false' ]; then
    pip install twine
    cd $TRAVIS_BUILD_DIR/lib/python
    make install
    twine upload dist/frugal-*.tar.gz
else
    echo "Not uploading to pypi as this isnt master branch"
fi
