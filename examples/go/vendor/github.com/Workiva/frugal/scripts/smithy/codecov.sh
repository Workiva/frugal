#!/bin/bash

# Takes a filepath as $1 and codecov flag as $2

# Example usage:
#   codecov.sh $FRUGAL_HOME/lib/java/target/site/jacoco/jacoco.xml java_library

if [[ -z $1 ]] || [[ -z $2 ]] ; then
    echo "Codecov.sh requires both a filepath and a flag"
    echo "Codecov.sh should be called as `codecov.sh filepath flag`"
    exit 1
fi

if [ -z "$GIT_BRANCH" ]
then
	echo "GIT_BRANCH environment variable not set, skipping codecov push"
else
	TRACKING_REMOTE="$(git for-each-ref --format='%(upstream:short)' $(git symbolic-ref -q HEAD) | cut -d'/' -f1 | xargs git ls-remote --get-url | cut -d':' -f2 | sed 's/.git$//')"
    bash <(curl -s https://codecov.workiva.net/bash) -t $CODECOV_TOKEN -r $TRACKING_REMOTE -f $1 -F $2 || echo "ERROR: Codecov failed to upload reports."

fi
