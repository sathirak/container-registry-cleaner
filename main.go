package main

import (
	"strconv"

	"github.com/sathirak/container-registry-cleaner/internal/cr"
	"github.com/sathirak/container-registry-cleaner/internal/types"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	registry := githubactions.GetInput("registry")
	repoName := githubactions.GetInput("name")
	if repoName == "" {
		githubactions.Fatalf("missing 'name'")
	}

	username := githubactions.GetInput("username")
	password := githubactions.GetInput("password")
	maxImagesStr := githubactions.GetInput("max-images")
	if maxImagesStr == "" {
		githubactions.Fatalf("missing 'max-images'")
	}
	maxImages, err := strconv.Atoi(maxImagesStr)
	if err != nil || maxImages < 0 {
		githubactions.Fatalf("invalid 'max-images': %v", err)
	}

	config := types.Config{
		Registry:  registry,
		RepoName:  repoName,
		Username:  username,
		Password:  password,
		MaxImages: maxImages,
	}

	if err := cr.CleanRegistry(config); err != nil {
		githubactions.Fatalf("error cleaning registry: %v", err)
	}
}
