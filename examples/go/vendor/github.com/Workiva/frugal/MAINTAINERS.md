## Release

To update the frugal version for release, use the python script provided in the
`scripts` directory. Note: You must be in the root directory of frugal.

Make sure you have the latest `develop`:

```bash
    $ git checkout develop && git pull
```

Checkout branch labeled with the release version:

```bash
    $ git checkout -b release_1_14_0
```

Note: The version for the release branch is determined via
[semantic versioning](http://semver.org/) based on the apis exposed in the
`master` branch.

Update version with python update script, commit, and push:

```bash
    $ python scripts/update.py --version=1.14.0
    $ git commit -m "Bump version to 1.14.0"
    $ git push origin release_1_14_0
```

From this point on, only bug fixes may be merged to the release branch.
When the release candidate is ready for release, open PR to `master` with label:

```
    MSG-161 RELEASE frugal 1.14.0
```



Where the MSG (or sometimes RM) ticket is the one tracked by MARV. Note: If the
version you are releasing doesn’t match the version tracked by MARV, double
click the version in MARV and change it to the appropriate value before you
open the PR.

Get at least two +1’s, merge and wait for Rosie to tag the release in github.
Download the smithy release artifact from the release build (e.g.  1.14.0)
rename to frugal-1.14.0-linux-amd64 and drag and drop the binary to the release
(e.g. 1.14.0) - note: you’ll have to click `Edit` on the release.
