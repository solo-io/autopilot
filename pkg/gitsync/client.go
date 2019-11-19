package gitsync
//
//import (
//	"context"
//	"fmt"
//	"github.com/pkg/errors"
//	"gopkg.in/src-d/go-git.v4"
//	"io"
//	"os"
//)
//
//type gitConfig struct {
//	url, repo, owner, branch string
//}
//
//func NewGitSyncer(url, repo, owner, branch, localDir string, progressWriter io.Writer) (*gitSyncer, error) {
//	cfg := gitConfig{url, repo, owner, branch}
//
//	var r *git.Repository
//
//	// if directory exists, simply open it and return
//	if _, err := os.Stat(localDir); err == nil {
//		r, err = git.PlainOpen(localDir)
//		if err != nil {
//			return nil, errors.Wrapf(err, "opening local repo")
//		}
//	} else {
//		// clone it
//		r, err = git.PlainClone(localDir, false, &git.CloneOptions{
//			SingleBranch: true,
//			Progress: progressWriter,
//		})
//		if err != nil {
//			return nil, errors.Wrapf(err, "cloning repo")
//		}
//	}
//
//	return &gitSyncer{
//		repo: r,
//		cfg:  cfg,
//	}, nil
//}
//
//func (cfg gitConfig) GetSyncer() (*gitSyncer, error) {
//
//}
//
//type gitSyncer struct {
//	cfg  gitConfig
//	repo *git.Repository
//}
//
//// Initialize the local git repo that will be used to stage, push, and commit changes
//func (s *gitSyncer) initRepo(localDir string) error {
//	// if directory exists, simply open it and return
//	if _, err := os.Stat(localDir); err == nil {
//		r, err := git.PlainOpen(localDir)
//		if err != nil {
//			return errors.Wrapf(err, "opening local repo")
//		}
//		s.repo = r
//		return nil
//	}
//
//	// clone the existing repository.
//	r, err := git.PlainClone(localDir, false, &git.CloneOptions{
//		u
//	})
//	if err != nil {
//		return errors.Wrapf(err, "opening local repo")
//	}
//
//	r, err := git.PlainClone(directory, false, &git.CloneOptions{
//		URL:               url,
//		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
//	})
//
//}
//
//func (s *gitSyncer) SyncToGit(ctx context.Context) {
//
//}
//
//func refName(branch string) string {
//	return fmt.Sprintf("refs/heads/%s", branch)
//}
