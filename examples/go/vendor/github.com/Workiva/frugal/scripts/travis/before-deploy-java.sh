#!/usr/bin/env bash
if [ "$TRAVIS_BRANCH" = 'master' ] && [ "$TRAVIS_PULL_REQUEST" == 'false' ]; then
    openssl aes-256-cbc -K $encrypted_e306f8772fe5_key \
        -iv $encrypted_e306f8772fe5_iv \
        -in $TRAVIS_BUILD_DIR/scripts/travis/codesigning.asc.enc \
        -out $TRAVIS_BUILD_DIR/scripts/travis/codesigning.asc -d
    gpg --fast-import $TRAVIS_BUILD_DIR/scripts/travis/codesigning.asc
else
    echo "Not deploying as this isnt master branch"
fi
