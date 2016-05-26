#!/bin/bash

if [ -z "$GIT_BRANCH" ]
then
	echo "GIT_BRANCH environment variable not set, skipping codecov push"
else
	bash <(curl -s https://codecov.workiva.net/bash) -u https://codecov.workiva.net -t $CODECOV_TOKEN -B $GIT_BRANCH -r Workiva/frugal-dart -f "coverage/coverage.lcov"
fi
