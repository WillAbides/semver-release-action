package release

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v53/github"
	"github.com/willabides/release-train-action/v3/internal"
	"github.com/willabides/release-train-action/v3/internal/next"
	"github.com/willabides/release-train-action/v3/internal/prev"
	"golang.org/x/mod/modfile"
)

type Runner struct {
	CheckoutDir    string
	Ref            string
	GithubToken    string
	CreateTag      bool
	CreateRelease  bool
	V0             bool
	TagPrefix      string
	InitialTag     string
	PrereleaseHook string
	GoModFiles     []string
	Repo           string
	PushRemote     string
	TempDir        string
	ReleaseRefs    []string
	GithubClient   internal.GithubClient
}

func (o *Runner) releaseNotesFile() string {
	return filepath.Join(o.TempDir, "release-notes")
}

func (o *Runner) releaseTargetFile() string {
	return filepath.Join(o.TempDir, "release-target")
}

var modVersionRe = regexp.MustCompile(`v\d+$`)

type Result struct {
	PreviousRef           string               `json:"previous-ref"`
	PreviousVersion       string               `json:"previous-version"`
	FirstRelease          bool                 `json:"first-release"`
	ReleaseVersion        *semver.Version      `json:"release-version,omitempty"`
	ReleaseTag            string               `json:"release-tag,omitempty"`
	ChangeLevel           internal.ChangeLevel `json:"change-level"`
	CreatedTag            bool                 `json:"created-tag,omitempty"`
	CreatedRelease        bool                 `json:"created-release,omitempty"`
	PrereleaseHookOutput  string               `json:"prerelease-hook-output"`
	PrereleaseHookAborted bool                 `json:"prerelease-hook-aborted"`
}

func (o *Runner) Next(ctx context.Context) (*Result, error) {
	ref := o.Ref
	if o.Ref == "" {
		ref = "HEAD"
	}
	head, err := runCmd(o.CheckoutDir, nil, "git", "rev-parse", ref)
	if err != nil {
		return nil, err
	}
	head = strings.TrimSpace(head)
	prevOpts := prev.Options{
		Head:     head,
		RepoDir:  o.CheckoutDir,
		Prefixes: []string{o.TagPrefix},
	}
	prevRef, err := prev.GetPrevTag(ctx, &prevOpts)
	if err != nil {
		return nil, err
	}
	firstRelease := prevRef == ""
	if firstRelease {
		result := Result{
			FirstRelease: true,
			ReleaseTag:   o.InitialTag,
			ChangeLevel:  internal.ChangeLevelNoChange,
		}
		if o.InitialTag != "" {
			result.ReleaseVersion, err = semver.NewVersion(strings.TrimPrefix(o.InitialTag, o.TagPrefix))
			if err != nil {
				return nil, err
			}
		}
		return &result, nil
	}
	prevVersion, err := semver.NewVersion(strings.TrimPrefix(prevRef, o.TagPrefix))
	if err != nil {
		return nil, err
	}

	maxBump := internal.ChangeLevelMajor
	if o.V0 {
		maxBump = internal.ChangeLevelMinor
		if prevVersion.Major() != 0 {
			return nil, fmt.Errorf("v0 flag is set, but previous version %q has major version > 0", prevVersion.String())
		}
	}

	result := Result{
		PreviousRef:     prevRef,
		PreviousVersion: prevVersion.String(),
	}
	var nextRes *next.Result
	nextRes, err = next.GetNext(ctx, &next.Options{
		Repo:         o.Repo,
		GithubClient: o.GithubClient,
		PrevVersion:  prevVersion.String(),
		Base:         prevRef,
		Head:         head,
		MaxBump:      maxBump.String(),
	})
	if err != nil {
		return nil, err
	}
	result.ReleaseVersion = &nextRes.NextVersion
	result.ReleaseTag = o.TagPrefix + nextRes.NextVersion.String()
	result.ChangeLevel = nextRes.ChangeLevel
	return &result, nil
}

func (o *Runner) runGoValidation(modFile string, result *Result) error {
	mfPath := filepath.Join(o.CheckoutDir, filepath.FromSlash(modFile))
	content, err := os.ReadFile(mfPath)
	if err != nil {
		return err
	}
	mf, err := modfile.ParseLax(mfPath, content, nil)
	if err != nil {
		return err
	}
	sv := result.ReleaseVersion
	major := int(sv.Major())
	wantM := ""
	if major > 1 {
		wantM = fmt.Sprintf("v%d", major)
	}
	m := modVersionRe.FindString(mf.Module.Mod.Path)
	if m != wantM {
		return fmt.Errorf("module %s has version suffix %q, want %q", mf.Module.Mod.Path, m, wantM)
	}
	return nil
}

func (o *Runner) repoOwner() string {
	return strings.SplitN(o.Repo, "/", 2)[0]
}

func (o *Runner) repoName() string {
	return strings.SplitN(o.Repo, "/", 2)[1]
}

func (o *Runner) getReleaseTarget() (string, error) {
	targetFile := o.releaseTargetFile()
	targetInfo, err := os.Stat(targetFile)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	target := ""
	if err == nil && !targetInfo.IsDir() {
		content, e := os.ReadFile(o.releaseTargetFile())
		if e != nil {
			return "", e
		}
		target = strings.TrimSpace(string(content))
	}
	if target == "" {
		return o.Ref, nil
	}
	return target, nil
}

func (o *Runner) getReleaseNotes(ctx context.Context, result *Result) (string, error) {
	notesInfo, err := os.Stat(o.releaseNotesFile())
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil && !notesInfo.IsDir() {
		content, e := os.ReadFile(o.releaseNotesFile())
		if e != nil {
			return "", e
		}
		return string(content), nil
	}
	// first release is empty by default
	if result.FirstRelease {
		return "", nil
	}
	return o.GithubClient.GenerateReleaseNotes(ctx, o.repoOwner(), o.repoName(), &github.GenerateNotesOptions{
		TagName:         result.ReleaseTag,
		PreviousTagName: &result.PreviousRef,
	})
}

func (o *Runner) Run(ctx context.Context) (*Result, error) {
	createTag := o.CreateTag
	release := o.CreateRelease
	if release {
		createTag = true
	}
	if len(o.ReleaseRefs) > 0 && !gitNameRev(o.CheckoutDir, o.Ref, o.ReleaseRefs) {
		createTag = false
		release = false
	}
	shallow, err := runCmd(o.CheckoutDir, nil, "git", "rev-parse", "--is-shallow-repository")
	if err != nil {
		return nil, err
	}
	if shallow == "true" {
		return nil, fmt.Errorf("shallow clones are not supported")
	}
	result, err := o.Next(ctx)
	if err != nil {
		return nil, err
	}

	if result.ReleaseVersion == nil || !createTag {
		return result, nil
	}
	if !result.FirstRelease && result.ChangeLevel == internal.ChangeLevelNoChange {
		return result, nil
	}

	runEnv := map[string]string{
		"RELEASE_VERSION":    result.ReleaseVersion.String(),
		"RELEASE_TAG":        result.ReleaseTag,
		"PREVIOUS_VERSION":   result.PreviousVersion,
		"FIRST_RELEASE":      fmt.Sprintf("%t", result.FirstRelease),
		"GITHUB_TOKEN":       o.GithubToken,
		"RELEASE_NOTES_FILE": o.releaseNotesFile(),
		"RELEASE_TARGET":     o.releaseTargetFile(),
	}

	result.PrereleaseHookOutput, result.PrereleaseHookAborted, err = runPrereleaseHook(o.CheckoutDir, runEnv, o.PrereleaseHook)
	if err != nil {
		return nil, err
	}
	if result.PrereleaseHookAborted {
		return result, nil
	}

	for _, mf := range o.GoModFiles {
		err = o.runGoValidation(mf, result)
		if err != nil {
			return nil, err
		}
	}

	target, err := o.getReleaseTarget()
	if err != nil {
		return nil, err
	}

	_, err = runCmd(o.CheckoutDir, nil, "git", "tag", result.ReleaseTag, target)
	if err != nil {
		return nil, err
	}

	_, err = runCmd(o.CheckoutDir, nil, "git", "push", o.PushRemote, result.ReleaseTag)
	if err != nil {
		return nil, err
	}

	result.CreatedTag = true

	if !release {
		return result, nil
	}

	releaseNotes, err := o.getReleaseNotes(ctx, result)
	if err != nil {
		return nil, err
	}

	prerelease := result.ReleaseVersion.Prerelease() != ""
	err = o.GithubClient.CreateRelease(ctx, o.repoOwner(), o.repoName(), &github.RepositoryRelease{
		TagName:    &result.ReleaseTag,
		Name:       &result.ReleaseTag,
		Body:       &releaseNotes,
		MakeLatest: github.String("legacy"),
		Prerelease: &prerelease,
	})
	if err != nil {
		return nil, err
	}

	result.CreatedRelease = true

	return result, nil
}

func runPrereleaseHook(dir string, env map[string]string, hook string) (stdout string, abort bool, _ error) {
	if hook == "" {
		return "", false, nil
	}
	var stdoutBuf bytes.Buffer
	cmd := exec.Command("sh", "-c", hook)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Stdout = &stdoutBuf
	err := cmd.Run()
	if err != nil {
		exitErr := asExitErr(err)
		if exitErr != nil {
			err = errors.Join(err, errors.New(string(exitErr.Stderr)))
			if exitErr.ExitCode() == 10 {
				return stdoutBuf.String(), true, nil
			}
		}
		return "", false, err
	}
	return stdoutBuf.String(), false, nil
}

func runCmd(dir string, env map[string]string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	out, err := cmd.Output()
	if err != nil {
		exitErr := asExitErr(err)
		if exitErr != nil {
			err = errors.Join(err, errors.New(string(exitErr.Stderr)))
		}
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func asExitErr(err error) *exec.ExitError {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr
	}
	return nil
}

func gitNameRev(dir, commitish string, refs []string) bool {
	args := []string{"name-rev", commitish, "--no-undefined"}
	for _, ref := range refs {
		args = append(args, "--refs", ref)
	}
	_, err := runCmd(dir, nil, "git", args...)
	return err == nil
}
