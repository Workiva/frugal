#!/usr/bin/env bash

GREEN='\e[1;32m'
RESET='\e[0m'

echo
echo -e $GREEN"Build information"
echo -e "--------------------------------------------------------------------------------"$RESET

echo "go env"
go env | sed "s/^/    /"
echo

echo "gitconfig"
git config -l | sed "s/^/    /"
echo

set -e

which godep > /dev/null || {
    go get github.com/tools/godep
}

cd $GOPATH/src/github.com/Workiva/frugal

echo "Pinning go dependencies."
godep restore

go install

# Run 
#go run main.go

echo "Go build successful."
