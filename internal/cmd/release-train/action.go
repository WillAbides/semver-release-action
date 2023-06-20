package releasetrain

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/willabides/release-train-action/v3/internal/actions"
	"github.com/willabides/release-train-action/v3/internal/orderedmap"
)

const (
	actionBoolSuffix  = "Only literal 'true' will be treated as true."
	actionSliceSuffix = "One value per line."
)

const (
	inputCheckPRLabels  = "check-pr-labels"
	inputCheckoutDir    = "checkout-dir"
	inputRef            = "ref"
	inputGithubToken    = "github-token"
	inputCreateTag      = "create-tag"
	inputCreateRelease  = "create-release"
	inputTagPrefix      = "tag-prefix"
	inputV0             = "v0"
	inputInitialTag     = "initial-release-tag"
	inputPreReleaseHook = "pre-release-hook"
	inputValidateGoMod  = "validate-go-module"
	inputReleaseRefs    = "release-refs"
	inputNoRelease      = "no-release"
	inputLabels         = "labels"
)

const (
	outputPreviousRef           = "previous-ref"
	outputPreviousVersion       = "previous-version"
	outputFirstRelease          = "first-release"
	outputReleaseVersion        = "release-version"
	outputReleaseTag            = "release-tag"
	outputChangeLevel           = "change-level"
	outputCreatedTag            = "created-tag"
	outputCreatedRelease        = "created-release"
	outputPreReleaseHookOutput  = "pre-release-hook-output"
	outputPreReleaseHookAborted = "pre-release-hook-aborted"
)

func getAction(kongCtx *kong.Context) *actions.CompositeAction {
	vars := kongCtx.Model.Vars()
	getVar := func(name string) string {
		val, ok := vars[name]
		if !ok {
			panic(fmt.Sprintf("variable %s not found", name))
		}
		return val
	}
	inputs := orderedmap.NewOrderedMap(
		orderedmap.Pair(inputCheckPRLabels, actions.Input{
			Description: `Instead of releasing, check that the PR has a label indicating the type of change.` +
				"\n\n" + actionBoolSuffix,
			Default: "${{ github.event_name == 'pull_request' }}",
		}),

		orderedmap.Pair(inputLabels, actions.Input{
			Description: getVar("label_help") + "\n" + actionSliceSuffix,
		}),

		orderedmap.Pair(inputCheckoutDir, actions.Input{
			Description: getVar("checkout_dir_help"),
			Default:     "${{ github.workspace }}",
		}),

		orderedmap.Pair(inputRef, actions.Input{
			Description: getVar("ref_help"),
			Default:     "${{ github.ref }}",
		}),

		orderedmap.Pair(inputGithubToken, actions.Input{
			Description: getVar("github_token_help"),
			Default:     "${{ github.token }}",
		}),

		orderedmap.Pair(inputCreateTag, actions.Input{
			Description: getVar("create_tag_help") + "\n\n" + actionBoolSuffix,
		}),

		orderedmap.Pair(inputCreateRelease, actions.Input{
			Description: getVar("create_release_help") + "\n\n" + actionBoolSuffix,
		}),

		orderedmap.Pair(inputTagPrefix, actions.Input{
			Description: getVar("tag_prefix_help"),
			Default:     vars["tag_prefix_default"],
		}),

		orderedmap.Pair(inputV0, actions.Input{
			Description: getVar("v0_help") + "\n\n" + actionBoolSuffix,
		}),

		orderedmap.Pair(inputInitialTag, actions.Input{
			Description: getVar("initial_tag_help"),
			Default:     vars["initial_tag_default"],
		}),

		orderedmap.Pair(inputPreReleaseHook, actions.Input{
			Description: getVar("pre_release_hook_help"),
		}),

		orderedmap.Pair(inputValidateGoMod, actions.Input{
			Description: getVar("go_mod_file_help") + "\n" + actionSliceSuffix,
		}),

		orderedmap.Pair(inputReleaseRefs, actions.Input{
			Description: getVar("release_ref_help") + "\n" + actionSliceSuffix,
		}),

		orderedmap.Pair(inputNoRelease, actions.Input{
			Description: `
If set to true, this will be a no-op. This is useful for creating a new repository or branch that isn't ready for
release yet.` + "\n\n" + actionBoolSuffix,
		}),
	)
	releaseStepEnv := orderedmap.NewOrderedMap(
		orderedmap.Pair("GITHUB_REPOSITORY", "${{ github.repository }}"),
	)
	for inputPair := inputs.Oldest(); inputPair != nil; inputPair = inputPair.Next() {
		envName := fmt.Sprintf("INPUT_%s", strings.ToUpper(inputPair.Key))
		envName = strings.ReplaceAll(envName, "-", "_")
		val := fmt.Sprintf("${{ inputs.%s }}", inputPair.Key)
		releaseStepEnv.AddPairs(orderedmap.Pair(envName, val))
	}

	releaseOutput := func(s string) string {
		return fmt.Sprintf("${{ steps.release.outputs.%s }}", s)
	}

	outputs := orderedmap.NewOrderedMap(
		orderedmap.Pair(outputPreviousRef, actions.CompositeOutput{
			Value:       releaseOutput(outputPreviousRef),
			Description: "A git ref pointing to the previous release, or the current ref if no previous release can be found.",
		}),

		orderedmap.Pair(outputPreviousVersion, actions.CompositeOutput{
			Value:       releaseOutput(outputPreviousVersion),
			Description: "The previous version on the release branch.",
		}),

		orderedmap.Pair(outputFirstRelease, actions.CompositeOutput{
			Value:       releaseOutput(outputFirstRelease),
			Description: "Whether this is the first release on the release branch. Either \"true\" or \"false\".",
		}),

		orderedmap.Pair(outputReleaseVersion, actions.CompositeOutput{
			Value:       releaseOutput(outputReleaseVersion),
			Description: "The version of the new release. Empty if no release is called for.",
		}),

		orderedmap.Pair(outputReleaseTag, actions.CompositeOutput{
			Value:       releaseOutput(outputReleaseTag),
			Description: "The tag of the new release. Empty if no release is called for.",
		}),

		orderedmap.Pair(outputChangeLevel, actions.CompositeOutput{
			Value:       releaseOutput(outputChangeLevel),
			Description: "The level of change in the release. Either \"major\", \"minor\", \"patch\" or \"none\".",
		}),

		orderedmap.Pair(outputCreatedTag, actions.CompositeOutput{
			Value:       releaseOutput(outputCreatedTag),
			Description: "Whether a tag was created. Either \"true\" or \"false\".",
		}),

		orderedmap.Pair(outputCreatedRelease, actions.CompositeOutput{
			Value:       releaseOutput(outputCreatedRelease),
			Description: "Whether a release was created. Either \"true\" or \"false\".",
		}),

		orderedmap.Pair(outputPreReleaseHookOutput, actions.CompositeOutput{
			Value:       releaseOutput(outputPreReleaseHookOutput),
			Description: "The stdout of the pre_release_hook. Empty if pre_release_hook is not set or if the hook returned an exit other than 0 or 10.",
		}),

		orderedmap.Pair(outputPreReleaseHookAborted, actions.CompositeOutput{
			Value:       releaseOutput(outputPreReleaseHookAborted),
			Description: "Whether pre_release_hook issued an abort by exiting 10. Either \"true\" or \"false\".",
		}),
	)

	releaseStep := actions.CompositeStep{
		Id:               "release",
		Shell:            "sh",
		WorkingDirectory: "${{ inputs.checkout-dir }}",
		Env:              releaseStepEnv,
		Run: `ACTION_DIR="${{ github.action_path }}"
if [ -z "$RELEASE_TRAIN_BIN" ]; then
  "$ACTION_DIR"/script/bindown -q install release-train --allow-missing-checksum
  RELEASE_TRAIN_BIN="$ACTION_DIR"/bin/release-train
fi

"$RELEASE_TRAIN_BIN" --action`,
	}

	return &actions.CompositeAction{
		Name:        "release-train",
		Description: "release-train keeps a-rollin' down to San Antone",
		Branding: &actions.Branding{
			Icon:  "send",
			Color: "yellow",
		},
		Inputs:  inputs,
		Outputs: outputs,
		Runs: actions.CompositeRuns{
			Using: "composite",
			Steps: []actions.CompositeStep{releaseStep},
		},
	}
}
