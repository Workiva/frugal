#!/usr/bin/env bash
if [ "$TRAVIS_BRANCH" = 'master' ] && [ "$TRAVIS_PULL_REQUEST" == 'false' ] && [ "$TRAVIS_REPO_SLUG" == 'Workiva/frugal' ]; then
    cd $TRAVIS_BUILD_DIR/lib/java && mvn clean deploy -P sign,build-extras \
        --settings $TRAVIS_BUILD_DIR/lib/java/.travis.settings.xml \
        -Dmaven.test.skip=true
fi
