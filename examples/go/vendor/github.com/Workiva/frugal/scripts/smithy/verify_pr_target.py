import re
import os
import sys


def _branch_matches_release_branch(branch_name):
    return re.search("release_[0-9]+.[0-9]+.[0-9]+", branch_name)


def _branch_matches_hotfix_branch(branch_name):
    return re.search("hotfix_[0-9]+.[0-9]+.[0-9]+", branch_name)


def _is_pull_request():
    return 'GIT_MERGE_BRANCH' in os.environ


def test_regex():
    assert(_branch_matches_release_branch('release_1.1.1'))
    assert(_branch_matches_release_branch('release_1234.1.1'))
    assert(not _branch_matches_release_branch('releases_1.1.1'))
    assert(not _branch_matches_release_branch('release_1.1'))
    assert(not _branch_matches_release_branch('release_1'))

    assert(_branch_matches_hotfix_branch('hotfix_1.1.1'))
    assert(_branch_matches_hotfix_branch('hotfix_1234.1.1'))
    assert(not _branch_matches_hotfix_branch('hotfixs_1.1.1'))
    assert(not _branch_matches_hotfix_branch('hotfix_1.1'))
    assert(not _branch_matches_hotfix_branch('hotfix_1'))

    assert(_allow_master_merge('hotfix_1.1.1'))
    assert(_allow_master_merge('hotfix_1234.1.1'))
    assert(not _allow_master_merge('hotfixs_1.1.1'))
    assert(not _allow_master_merge('hotfix_1.1'))
    assert(not _allow_master_merge('hotfix_1'))
    

def _allow_master_merge(branch_name):
    return _branch_matches_release_branch(branch_name) or _branch_matches_hotfix_branch(branch_name)


def main():
    """
    Verifies when making a PR, the target branch is not master unless the
    current branch matches a regex for a release PR
    """

    if _is_pull_request():

        merge_branch = os.environ['GIT_MERGE_BRANCH']
        cur_branch = os.environ['GIT_BRANCH']

        if merge_branch == 'master' and not _allow_master_merge(cur_branch):
            print('ERROR: Your branch:{cur_branch} does not appear to be a release PR, but was made against master instead of develop.'.format(cur_branch=cur_branch))
            sys.exit(1)


main()
