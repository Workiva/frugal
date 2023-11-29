import re
import os
import sys


def _branch_matches_release_branch(branch_name):
    return re.search('refs/heads/release_v?[0-9]+.[0-9]+.[0-9]+', branch_name)


def _branch_matches_hotfix_branch(branch_name):
    return re.search('refs/heads/hotfix_[0-9]+.[0-9]+.[0-9]+', branch_name)


def _branch_matches_docs_branch(branch_name):
    return re.search('refs/heads/docs_.*', branch_name)


def _is_pull_request():
    return 'GIT_MERGE_BRANCH' in os.environ


def test_regex():
    assert(_allow_master_merge('refs/heads/release_1.1.1'))
    assert(_allow_master_merge('refs/heads/release_1234.1.1'))
    assert(not _allow_master_merge('refs/heads/releases_1.1.1'))
    assert(not _allow_master_merge('refs/heads/release_1.1'))
    assert(not _allow_master_merge('refs/heads/release_1'))

    assert(_allow_master_merge('refs/heads/hotfix_1.1.1'))
    assert(_allow_master_merge('refs/heads/hotfix_1234.1.1'))
    assert(not _allow_master_merge('refs/heads/hotfixes_1.1.1'))
    assert(not _allow_master_merge('refs/heads/hotfix_1.1'))
    assert(not _allow_master_merge('refs/heads/hotfix_1'))

    assert(_allow_master_merge('refs/heads/docs_update_upgrade_doc'))
    assert(not _allow_master_merge('refs/heads/docsupdate_upgrade_doc'))


def _allow_master_merge(branch_name):
    return _branch_matches_release_branch(branch_name) or \
        _branch_matches_hotfix_branch(branch_name) or \
        _branch_matches_docs_branch(branch_name)


def main():
    """
    Verifies when making a PR, the target branch is not master unless the
    current branch matches a regex for a release PR
    """

    if _is_pull_request():

        merge_branch = os.environ['GIT_MERGE_BRANCH']
        cur_branch = os.environ['GIT_BRANCH']

        if merge_branch == 'refs/heads/master' and not _allow_master_merge(cur_branch):
            print('ERROR: Your branch:{cur_branch} does not appear to be a '
                  'release or documentation PR, but was made against '
                  'master instead of develop.'.format(cur_branch=cur_branch))
            sys.exit(1)


if __name__ == '__main__':

    # TEST=true python scripts/smithy/verify_pr_target.py
    if 'TEST' in os.environ:
        test_regex()
    else:
        main()
