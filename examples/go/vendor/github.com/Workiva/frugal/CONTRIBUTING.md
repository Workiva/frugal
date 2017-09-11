Contributing to Frugal
=======================

Creating Pull Requests
----------------------

 * Write [good commit messages](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).
 * Prefix the commit message area impacted by the commit - compiler, language, etc, i.e. "Python: Fix annoying bug".
 * Branch off `develop`, PR back to `develop`.
 	* http://nvie.com/posts/a-successful-git-branching-model/

Reviewing Code
--------------

 - We require two +1s on the last commit of every PR before it is merged.
 - If you think a PR could affect security (or Aviary flags the PR as a
   security-related PR), it must also have a separate "security +1" on the last
   commit in addition to two +1s.
    - The security +1 can come from one of the devs who +1'd the PR in the first
      place.
 - To request code reviews from the team as a whole, include "@Workiva/messaging-pp" in your PR message.

Current Frugal Maintainers
------------------------------

 - Brian Shannan
 - Charlie Strawn
 - Steven Osborne
 - Tyler Rinnan
