# Contributing to Compono

Thanks for your interest in Compono. Please read this page before you put time into a contribution, so none of it is wasted.

## How Compono is developed

- **Built for Umono.** Compono is developed for [Umono](https://github.com/umono-cms/umono) and follows Umono's roadmap.
- **One developer.** Compono is developed by Yunus Emre Bulut alone. There is no team.
- **The process is private.** Decisions, specs and planning live in a private document. What is public is the code, the README, the issues and the discussions.
- **Built with AI.** The code is written by AI agents. The design, the decisions and the review are the developer's.

## How you can help

The most useful contributions are issues and discussions.

- **Bugs:** open an [issue](https://github.com/umono-cms/compono/issues/new). Include the input, the output you got and the output you expected.
- **Feedback, questions and ideas:** use [GitHub Discussions](https://github.com/umono-cms/compono/discussions).
- **Security vulnerabilities:** do not report them in public. Follow [SECURITY.md](SECURITY.md).

## HTML output bugs

Since Umono v0.9, Compono's HTML output is part of Umono's Theme API, and themes depend on it. While Umono is in 0.x, fixes that change the HTML output do not go into the main branch. They wait for Umono v1, so existing themes keep working. Security fixes are the exception.

Please report such a bug anyway. It is recorded and fixed for v1.

## Code contributions

Compono does not need many code contributions. If you still want to write code:

1. **Work on an assigned task only.** The task must be decided in an issue or a discussion and assigned to you. Deciding and assigning are the developer's call, because decisions are made in the private document and a change that has not been decided may not fit them.
2. **Pull requests for unassigned work are closed without review.** There are no exceptions, typo fixes included. Every change starts as an issue or a discussion.
3. **Declare your AI use.** Using AI is fine. The pull request description says which AI tools you used and what was written with them. Pull requests without this declaration are closed.
4. **Behavior changes come with tests.** This applies to the project's own changes too.

— Yunus Emre Bulut
