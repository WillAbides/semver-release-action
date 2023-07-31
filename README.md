# release-train

[Contributions welcome](./CONTRIBUTING.md).

**release-train** creates releases for every PR merge in your repository. No
magic commit message is required. You only need to label your PRs with the
appropriate labels and run release-train (either from the command line or a 
GitHub Action).

Release-train is inspired
by [semantic-release](https://github.com/semantic-release/semantic-release) and
has a few advantages for somebody with my biases:
- It doesn't require special commit messages, so you won't need to squash
  commits or ask contributors to rebase before merging.
- No need for npm or other package managers.
- No plugin configuration. Release-train has no plugins. You can do anything
  a plugin would do from the pre-release hook.

## Release steps

When run from a release branch, release-train follows these steps to publish a
release. Some options such as `--check-pr` will modify this behavior.

1. **Find the previous release tag** by searching backward through git history
   for the first tag formatted like `<prefix><semantic version>`. When no
   previous tag is found it uses `v0.0.0` or the value from `--initial-tag`.
2. **Find the next release version** by looking at the diff between HEAD and the
   previous release tag and inspecting the pull request where each commit was
   introduced. The previous version is incremented by the highest change level
   found.
3. **Run pre-release-hook**. This is where you can do things like validate the
   release, built release artifacts or generate a changelog.
4. **Create and push the new git tag** if `--create-tag` is set.
5. **Create a draft release** if `--create-release` is set. It starts as a draft
   to avoid publishing a release that doesn't have all the necessary artifacts
   yet.
6. **Upload release assets**. Any files written to `$ASSETS_DIR` will be
   uploaded as release assets.
7. **Publish the release**.
8. **Emit output** including release version, tag, change level, etc that you
   use in notifications or other actions.

## GitHub Action Configuration

See [action.md](./doc/action.md).

## Command Line

### Installation

#### Install with [bindown](https://github.com/WillAbides/bindown)

```shell
bindown template-source add release-train https://github.com/WillAbides/release-train/releases/latest/download/bindown.yaml
bindown dependency add release-train --source release-train
```

#### Install from go source

```shell
go install github.com/willabides/release-train/v3@latest
```

#### Download a release

Pick a release from
the [releases page](https://github.com/WillAbides/release-train/releases) and
download the appropriate binary for your platform.

### Usage

<!--- start usage output --->

```
Usage: release-train

Release every PR merge. No magic commit message required.

Flags:
  -h, --help                          Show context-sensitive help.
      --version
      --repo=STRING                   GitHub repository in the form of owner/repo.
      --check-pr=INT                  Operates as if the given PR has already been merged. Useful
                                      for making sure the PR is properly labeled. Skips tag and
                                      release.
      --label=<alias>=<label>;...     PR label alias in the form of "<alias>=<label>" where <label>
                                      is a canonical label.
  -C, --checkout-dir="."              The directory where the repository is checked out.
      --ref="HEAD"                    git ref.
      --create-tag                    Whether to create a tag for the release.
      --create-release                Whether to create a release. Implies create-tag.
      --draft                         Leave the release as a draft.
      --tag-prefix="v"                The prefix to use for the tag.
      --v0                            Assert that current major version is 0 and treat breaking
                                      changes as minor changes. Errors if the major version is not
                                      0.
      --initial-tag="v0.0.0"          The tag to use if no previous version can be found. Set to ""
                                      to cause an error instead.
      --pre-release-hook=<command>    *deprecated* Will be removed in a future release. Alias for
                                      pre-tag-hook.
      --release-ref=<branch>,...      Only allow tags and releases to be created from matching refs.
                                      Refs can be patterns accepted by git-show-ref. If undefined,
                                      any branch can be used.
      --push-remote="origin"          The remote to push tags to.
      --tempdir=STRING                The prefix to use with mktemp to create a temporary directory.
      --github-api-url="https://api.github.com"
                                      GitHub API URL.
      --output-format="json"          Output either json our GitHub action output.
      --debug                         Enable debug logging.
```

<!--- end usage output --->
