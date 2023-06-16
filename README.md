# release-train

Release-train is a command-line tool and GitHub action for creating releases based on Pull Request labels.

# Action

<!--- everything between the next line and the "end action doc" comment is generated by script/generate --->
<!--- start action doc --->

release-train keeps a-rollin' down to San Antone

## Inputs

### check_pr_labels

default: `${{ github.event_name == 'pull_request' }}`

Instead of releasing, check that the PR has a label indicating the type of change.

Only literal 'true' will be treated as true.

### checkout_dir

default: `${{ github.workspace }}`

The directory where the repository is checked out.

### ref

default: `${{ github.ref }}`

git ref.

### github_token

default: `${{ github.token }}`

The GitHub token to use for authentication. Must have `contents: write` permission if creating a release or tag.

### create_tag

Whether to create a tag for the release. Implies create-tag.

Only literal 'true' will be treated as true.

### create_release

Whether to create a release. Implies create-tag.

Only literal 'true' will be treated as true.

### tag_prefix

default: `v`

The prefix to use for the tag.

### v0

Assert that current major version is 0 and treat breaking changes as minor changes. 
Errors if the major version is not 0.


Only literal 'true' will be treated as true.

### initial_release_tag

default: `v0.0.0`

The tag to use if no previous version can be found. Set to "" to cause an error instead.

### pre_release_hook

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

The environment variables RELEASE_VERSION, RELEASE_TAG, PREVIOUS_VERSION, FIRST_RELEASE, GITHUB_TOKEN,
RELEASE_NOTES_FILE and RELEASE_TARGET will be set.


### validate_go_module

Validates that the name of the go module at the given path matches the major version of the release. For example,
validation will fail when releasing v3.0.0 when the module name is "my_go_module/v2".


### release_refs

Only allow tags and releases to be created from matching refs. Refs can be patterns accepted by git-show-ref. 
If undefined, any branch can be used.


Comma separated list of values without spaces.

### no_release

If set to true, this will be a no-op. This is useful for creating a new repository or branch that isn't ready for
release yet.

Only literal 'true' will be treated as true.

## Outputs

### previous_ref

A git ref pointing to the previous release, or the current ref if no previous release can be found.

### previous_version

The previous version on the release branch.

### first_release

Whether this is the first release on the release branch. Either "true" or "false".

### release_version

The version of the new release. Empty if no release is called for.

### release_tag

The tag of the new release. Empty if no release is called for.

### change_level

The level of change in the release. Either "major", "minor", "patch" or "no change".

### created_tag

Whether a tag was created. Either "true" or "false".

### created_release

Whether a release was created. Either "true" or "false".

### pre_release_hook_output

The stdout of the pre_release_hook. Empty if pre_release_hook is not set or if the hook returned an exit other than 0 or 10.

### pre_release_hook_aborted

Whether pre_release_hook issued an abort by exiting 10. Either "true" or "false".
<!--- end action doc --->
