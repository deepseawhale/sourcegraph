package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	// DO NOT CHANGE THIS. We need this to be stable to have stable git revhash.
	// Changing this would cause old links to 404 as previous revhashes would
	// be invalidated by version tags having a new revhash.
	stableGitCommitDate = "Thu Apr 8 14:24:52 2021 +0200"
)

type JvmPackagesArtifactSyncer struct {
	Config *schema.JvmPackagesConnection
}

var _ VCSSyncer = &JvmPackagesArtifactSyncer{}

func (s JvmPackagesArtifactSyncer) Type() string {
	return "jvm_packages"
}

// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
// error indicates there is a problem.
func (s JvmPackagesArtifactSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	dependencies, err := s.PackageDependencies(remoteURL.Path)
	if err != nil {
		return err
	}

	for _, dependency := range dependencies {
		sources, err := coursier.FetchSources(ctx, s.Config, dependency)
		if err != nil {
			return err
		}
		if len(sources) == 0 {
			return errors.Errorf("no sources.jar for dependency %s", dependency)
		}
	}
	return nil
}

// CloneCommand returns the command to be executed for cloning from remote.
func (s JvmPackagesArtifactSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, bareGitDirectory string) (*exec.Cmd, error) {
	err := os.MkdirAll(bareGitDirectory, 0755)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "git", "--bare", "init", "--initial-branch", "main")
	if err := runCommandInDirectory(ctx, cmd, bareGitDirectory); err != nil {
		return nil, err
	}

	if err := s.Fetch(ctx, remoteURL, GitDir(bareGitDirectory)); err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, "git", "--version"), nil
}

// Fetch does nothing for Maven packages because they are immutable and cannot be updated after publishing.
func (s JvmPackagesArtifactSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	dependencies, err := s.PackageDependencies(remoteURL.Path)
	if err != nil {
		return err
	}

	tags := make(map[string]bool)

	cmd := exec.CommandContext(ctx, "git", "tag")
	cmd.Dir = string(dir)
	out, err := runWith(ctx, cmd, false, nil)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}
		tags[line] = true
	}

	for i, dependency := range dependencies {
		if _, ok := tags[dependency.GitTagFromVersion()]; ok {
			continue
		}
		if err := s.gitPushDependencyTag(ctx, string(dir), dependency, i == 0); err != nil {
			log15.Error("error pushing dependency tag", "error", err, "tag", dependency)
			return err
		}
	}

	for tag := range tags {
		shouldBeRemoved := true
		for _, dependency := range dependencies {
			if dependency.GitTagFromVersion() == tag {
				shouldBeRemoved = false
				break
			}
		}

		if shouldBeRemoved {
			cmd := exec.CommandContext(ctx, "git", "tag", "-d", tag)
			if err := runCommandInDirectory(ctx, cmd, string(dir)); err != nil {
				log15.Error("failed to delete git tag", "error", err, "tag", tag)
				continue
			}
		}
	}

	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote.
func (s JvmPackagesArtifactSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
}

// PackageDependencies returns the list of JVM dependencies that belong to the given URL path.
// A URL maps to a single JVM package, which may contain multiple versions (one git tag per version).
func (s JvmPackagesArtifactSyncer) PackageDependencies(repoUrlPath string) (dependencies []reposource.Dependency, err error) {
	module, err := reposource.ParseMavenModule(repoUrlPath)
	if err != nil {
		return nil, err
	}
	for _, dependency := range s.Config.Maven.Artifacts {
		if module.MatchesDependencyString(dependency) {
			dependency, err := reposource.ParseMavenDependency(dependency)
			if err != nil {
				return nil, err
			}
			dependencies = append(dependencies, dependency)
		}
	}
	if len(dependencies) == 0 {
		return nil, errors.Errorf("no tracked dependencies for URL path %s", repoUrlPath)
	}
	reposource.SortDependencies(dependencies)
	return dependencies, nil
}

// gitPushDependencyTag pushes a git tag to the given bareGitDirectory path. The
// tag points to a commit that adds all sources of given dependency. When
// isMainBranch is true, the main branch of the bare git directory will also be
// updated to point to the same commit as the git tag.
func (s JvmPackagesArtifactSyncer) gitPushDependencyTag(ctx context.Context, bareGitDirectory string, dependency reposource.Dependency, isMainBranch bool) error {
	tmpDirectory, err := ioutil.TempDir("", "maven")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDirectory)

	paths, err := coursier.FetchSources(ctx, s.Config, dependency)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		return errors.Errorf("no sources.jar for dependency %s", dependency)
	}

	path := paths[0]

	cmd := exec.CommandContext(ctx, "git", "init", "--initial-branch", "main")
	if err := runCommandInDirectory(ctx, cmd, tmpDirectory); err != nil {
		return err
	}

	err = s.commitJar(ctx, dependency, tmpDirectory, path)
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", bareGitDirectory)
	if err := runCommandInDirectory(ctx, cmd, tmpDirectory); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "push", "origin", "--tags")
	if err := runCommandInDirectory(ctx, cmd, tmpDirectory); err != nil {
		return err
	}

	if isMainBranch {
		cmd = exec.CommandContext(ctx, "git", "push", "--force", "origin", "main", dependency.GitTagFromVersion())
		if err := runCommandInDirectory(ctx, cmd, tmpDirectory); err != nil {
			return err
		}
	}

	return nil
}

// commitJar creates a git commit in the given working directory that adds all the file contents of the given jar file.
// A `*.jar` file works the same way as a `*.zip` file, it can even be uncompressed with the `unzip` command-line tool.
func (s JvmPackagesArtifactSyncer) commitJar(ctx context.Context, dependency reposource.Dependency, workingDirectory, jarPath string) error {
	cmd := exec.CommandContext(ctx, "unzip", jarPath, "-d", "./")
	if err := runCommandInDirectory(ctx, cmd, workingDirectory); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(workingDirectory, "lsif-java.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContents, err := json.Marshal(&lsifJavaJson{
		Kind:         "maven",
		Jvm:          "8",
		Dependencies: []string{dependency.CoursierSyntax()},
	})
	if err != nil {
		return err
	}

	_, err = file.Write(jsonContents)
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = workingDirectory
	if err := runCommandInDirectory(ctx, cmd, workingDirectory); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", dependency.CoursierSyntax(), "--date", stableGitCommitDate)
	if err := runCommandInDirectory(ctx, cmd, workingDirectory); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "tag", "-m", dependency.CoursierSyntax(), dependency.GitTagFromVersion())
	if err := runCommandInDirectory(ctx, cmd, workingDirectory); err != nil {
		return err
	}

	return nil
}

func runCommandInDirectory(ctx context.Context, cmd *exec.Cmd, workingDirectory string) error {
	cmd.Dir = workingDirectory
	output, err := runWith(ctx, cmd, false, nil)
	if err != nil {
		return errors.Wrapf(err, "command %s failed with output %q", cmd.Args, string(output))
	}
	return nil
}

type lsifJavaJson struct {
	Kind         string   `json:"kind"`
	Jvm          string   `json:"jvm"`
	Dependencies []string `json:"dependencies"`
}
