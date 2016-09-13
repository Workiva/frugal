import re
import os
import sys


def _branch_matches_release_branch(branch_name):
    return re.search("release_[0-9]+.[0-9]+.[0-9]+", branch_name)


def _is_pull_request():
    return 'GIT_MERGE_BRANCH' in os.environ


def test_regex():
    assert(_branch_matches_release_branch('release_1.1.1'))
    assert(_branch_matches_release_branch('release_1234.1.1'))
    assert(not _branch_matches_release_branch('releases_1.1.1'))
    assert(not _branch_matches_release_branch('release_1.1'))
    assert(not _branch_matches_release_branch('release_1'))
    

def main():
    """
    Verifies when making a PR, the target branch is not master unless the
    current branch matches a regex for a release PR
    """

    if _is_pull_request():

        merge_branch = os.environ['GIT_MERGE_BRANCH']
        cur_branch = os.environ['GIT_BRANCH']

        if merge_branch == 'master' and not _branch_matches_release_branch(cur_branch):
            print('ERROR: Your branch:{cur_branch} does not appear to be a release PR, but was made against master'.format(cur_branch=cur_branch))
            sys.exit(1)


main()
