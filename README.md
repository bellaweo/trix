[![status-badge](https://ci.codeberg.org/api/badges/meh/trix/status.svg)](https://ci.codeberg.org/meh/trix)

# trix

A matrix cli for performing one-off tasks.

The cli is desigend to be mostly self-documenting. To see the cmd line options, run `trix help`.

Currently, the cli supports sending encrypted messages to a matrix room. The user needs to already exist on the matrix host and needs permission to join the the matrix room. The primary use-case for this project is in scripts to send notifications to a matrix room.

Current releases in this repo are verified to work on debian/ubuntu flavor linux hosts. The libolm C libraries must be installed onto the host to support matrix encryption. I haven't tested this on other linux falvors or macos yet.

# development

An integration test suite exists in this repo which is managed by [Earthly](https://earthly.dev). Once you have earthly installed, the Earthfile in the root of the repo has a +test target which bootstraps an isolated matrix server and tests the trix binary against it.

The Earthfile +all target will build the trix binary, run the integration tests, and create a local trix artifact.

Tests can be run in debug mode by providing the DEBUG ARG to the earthly command. For example, `earthly --build-arg DEBUG=true +all`

# Give Up GitHub

This project has given up GitHub.  ([See Software Freedom Conservancy's *Give Up  GitHub* site for details](https://GiveUpGitHub.org).)

This project is mirrored to GitHub. It is actually located at  [Codeberg](https://codeberg.org/meh/trix).

Any use of this project's code by GitHub Copilot, past or present, is done without our permission.  We do not consent to GitHub's use of this project's code in Copilot.

