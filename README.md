# release-train

Release-train is a command-line tool and GitHub action for creating releases based on Pull Request labels.

# Action

<!--- everything between the next line and the "end action doc" comment is generated by script/generate --->
<!--- start action doc --->

release-train keeps a-rollin' down to San Antone

## Inputs

### repo

default: `${{ github.repository }}`

Github repository in the form of owner/repo.

### check-pr

default: `${{ github.event.number }}`

Check that this PR number is properly labeled. Skips tag and release when set.

### labels

PR label alias in the form of "<alias>=<label>" where <label> is a canonical label.

Accepts multiple values. One value per line.

### checkout-dir

default: `${{ github.workspace }}`

The directory where the repository is checked out.

### ref

default: `${{ github.ref }}`

git ref.

### github-token

default: `${{ github.token }}`

The GitHub token to use for authentication. Must have `contents: write` permission if creating a release or tag.

### create-tag

Whether to create a tag for the release. Implies create-tag.

Only literal 'true' will be treated as true.

### create-release

Whether to create a release. Implies create-tag.

Only literal 'true' will be treated as true.

### tag-prefix

default: `v`

The prefix to use for the tag.

### v0

Assert that current major version is 0 and treat breaking changes as minor changes. 
Errors if the major version is not 0.

Only literal 'true' will be treated as true.

### initial-release-tag

default: `v0.0.0`

The tag to use if no previous version can be found. Set to "" to cause an error instead.

### pre-release-hook

Command to run before creating the release. You may abort the release by exiting with a non-zero exit code.

Exit code 0 will continue the release. Exit code 10 will skip the release without error. Any other exit code will
abort the release with an error.

You may provide custom release notes by writing to the file at `$RELEASE_NOTES_FILE`:

```
  echo "my release notes" > "$RELEASE_NOTES_FILE"
```

You can update the git ref to be released by writing it to the file at `$RELEASE_TARGET`:

```
  # ... update some files ...
  git commit -am "prepare release $RELEASE_TAG"
  echo "$(git rev-parse HEAD)" > "$RELEASE_TARGET"
```

If you create a tag named $RELEASE_TAG, it will be used as the release target instead of either HEAD or the value
written to $RELEASE_TARGET.

Any files written to $ASSETS_DIR will be uploaded as release assets.

The environment variables RELEASE_VERSION, RELEASE_TAG, PREVIOUS_VERSION, FIRST_RELEASE, GITHUB_TOKEN,
RELEASE_NOTES_FILE, RELEASE_TARGET and ASSETS_DIR will be set.

### validate-go-module

Validates that the name of the go module at the given path matches the major version of the release. For example,
validation will fail when releasing v3.0.0 when the module name is "my_go_module/v2".

Accepts multiple values. One value per line.

### release-refs

Only allow tags and releases to be created from matching refs. Refs can be patterns accepted by git-show-ref.
If undefined, any branch can be used.

Accepts multiple values. One value per line.

### tempdir

The prefix to use with mktemp to create a temporary directory.

### release-train-bin

Path to release-train binary. Only needed if you're using a custom release-train binary.

## Outputs

### previous-ref

A git ref pointing to the previous release, or the current ref if no previous release can be found.

### previous-version

The previous version on the release branch.

### first-release

Whether this is the first release on the release branch. Either "true" or "false".

### release-version

The version of the new release. Empty if no release is called for.

### release-tag

The tag of the new release. Empty if no release is called for.

### change-level

The level of change in the release. Either "major", "minor", "patch" or "none".

### created-tag

Whether a tag was created. Either "true" or "false".

### created-release

Whether a release was created. Either "true" or "false".

### pre-release-hook-output

The stdout of the pre_release_hook. Empty if pre_release_hook is not set or if the hook returned an exit other than 0 or 10.

### pre-release-hook-aborted

Whether pre_release_hook issued an abort by exiting 10. Either "true" or "false".
<!--- end action doc --->
