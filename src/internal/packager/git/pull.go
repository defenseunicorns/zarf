// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package git contains functions for interacting with git repositories.
package git

import (
	"errors"
	"fmt"

	"path/filepath"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// DownloadRepoToTemp clones or updates a repo into a temp folder to perform ephemeral actions (i.e. process chart repos).
func (g *Git) DownloadRepoToTemp(gitURL string) string {
	path, err := utils.MakeTempDir(config.CommonOptions.TempDirectory)
	if err != nil {
		message.Fatalf(err, "Unable to create tmpdir: %s", config.CommonOptions.TempDirectory)
	}
	// If downloading to temp, grab all tags since the repo isn't being
	// packaged anyway, and it saves us from having to fetch the tags
	// later if we need them

	err = g.pull(gitURL, path, "")
	return path
}

// Pull clones or updates a git repository into the target folder.
func (g *Git) Pull(gitURL, targetFolder string) (path string, err error) {
	repoName, err := g.TransformURLtoRepoName(gitURL)
	if err != nil {
		message.Errorf(err, "unable to pull the git repo at %s", gitURL)
		return "", err
	}

	path = targetFolder + "/" + repoName
	g.GitPath = path
	err = g.pull(gitURL, path, repoName)
	return path, err
}

// internal pull function that will clone/pull the latest changes from the git repo
func (g *Git) pull(gitURL, targetFolder string, repoName string) error {
	g.Spinner.Updatef("Processing git repo %s", gitURL)

	gitCachePath := targetFolder
	if repoName != "" {
		gitCachePath = filepath.Join(config.GetAbsCachePath(), filepath.Join(config.ZarfGitCacheDir, repoName))
	}

	matches := gitURLRegex.FindStringSubmatch(gitURL)
	idx := gitURLRegex.SubexpIndex

	if len(matches) == 0 {
		// Unable to find a substring match for the regex
		return fmt.Errorf("unable to get extract the repoName from the url %s", gitURL)
	}

	onlyFetchRef := matches[idx("atRef")] != ""
	gitURLNoRef := fmt.Sprintf("%s%s/%s%s", matches[idx("proto")], matches[idx("hostPath")], matches[idx("repo")], matches[idx("git")])

	repo, err := g.clone(gitCachePath, gitURLNoRef, onlyFetchRef)

	if err == git.ErrRepositoryAlreadyExists {

		// Pull the latest changes from the online repo
		message.Debug("Repo already cloned, pulling any upstream changes...")
		gitCred := utils.FindAuthForHost(gitURL)
		pullOptions := &git.PullOptions{
			RemoteName: onlineRemoteName,
			Auth:       &gitCred.Auth,
		}
		worktree, err := repo.Worktree()
		if err != nil {
			message.Debugf("unable to get the worktree for the repo: %s", gitURL)
			return err
		}
		err = worktree.Pull(pullOptions)
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			message.Debug("Repo already up to date")
		} else if err != nil {
			g.Spinner.Warnf("Not a valid git repo or unable to pull: %s", gitURL)
			return err
		}

		// NOTE: Since pull doesn't pull any new tags, we need to fetch them
		fetchOptions := git.FetchOptions{RemoteName: onlineRemoteName, Tags: git.AllTags}
		if err := g.fetch(gitCachePath, &fetchOptions); err != nil {
			return err
		}

	} else if err != nil {
		g.Spinner.Warnf("Not a valid git repo or unable to clone: %s", gitURL)
		return err
	}

	if gitCachePath != targetFolder {
		err = utils.CreatePathAndCopy(gitCachePath, targetFolder)
		if err != nil {
			return fmt.Errorf("unable to copy %s into %s: %#v", gitCachePath, targetFolder, err.Error())
		}
	}

	if onlyFetchRef {
		ref := matches[idx("ref")]

		// Identify the remote trunk branch name
		trunkBranchName := plumbing.NewBranchReferenceName("master")
		head, err := repo.Head()

		if err != nil {
			// No repo head available
			g.Spinner.Errorf(err, "Failed to identify repo head. Ref will be pushed to 'master'.")
		} else if head.Name().IsBranch() {
			// Valid repo head and it is a branch
			trunkBranchName = head.Name()
		} else {
			// Valid repo head but not a branch
			g.Spinner.Errorf(nil, "No branch found for this repo head. Ref will be pushed to 'master'.")
		}

		_, _ = g.removeLocalTagRefs()
		_, _ = g.removeLocalBranchRefs()
		_, _ = g.removeOnlineRemoteRefs()

		err = g.fetchRef(ref)
		if err != nil {
			return fmt.Errorf("not a valid reference or unable to fetch (%s): %#v", ref, err)
		}

		err = g.checkoutRefAsBranch(ref, trunkBranchName)
		return err
	}

	return nil
}
