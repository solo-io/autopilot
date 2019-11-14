package main

import (
	"github.com/solo-io/go-utils/githubutils"
)

const buildDir = "_output"
const repoOwner = "solo-io"
const repoName = "autopilot"

func main() {
	assets := []githubutils.ReleaseAssetSpec{
		{
			Name:       "ap-linux-amd64",
			ParentPath: buildDir,
			UploadSHA:  true,
		},
		{
			Name:       "ap-darwin-amd64",
			ParentPath: buildDir,
			UploadSHA:  true,
		},
		{
			Name:       "ap-windows-amd64.exe",
			ParentPath: buildDir,
			UploadSHA:  true,
		},
	}
	spec := githubutils.UploadReleaseAssetSpec{
		Owner:             repoOwner,
		Repo:              repoName,
		Assets:            assets,
		SkipAlreadyExists: true,
	}
	githubutils.UploadReleaseAssetCli(&spec)
}
