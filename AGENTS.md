Comments
--------

Prefer self-documenting code.

Do not add doc comments just because a Go symbol is exported.

Add comments only when they explain something non-obvious that the
code itself cannot express clearly.

Commit messages
---------------

Use ASCII only. Require an imperative subject starting with a capital
letter, without markup or a trailing period. Aim for at most 50 characters.

Add body paragraphs only as needed, separated from the subject by a blank
line and wrapped at 72 characters. Be concise; inline code and emphasis
markup are allowed in the body.

Change scope and commits
------------------------

Keep changes minimal and focused. Avoid unrelated fixes or style changes.
Put necessary refactoring, including test refactoring, in separate commits.

Make each commit self-contained and easy to review. Order commits so the
history explains the work without lengthy messages. Simple changes may
include tests; separate larger changes from their tests, and split
independent test changes into separate commits.
