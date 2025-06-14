package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

	fullRepo := repoName
	if registry != "" {
		fullRepo = fmt.Sprintf("%s/%s", registry, repoName)
	}

	ref, err := name.NewRepository(fullRepo)
	if err != nil {
		log.Fatal(err)
	}

	var opts []remote.Option
	if username != "" && password != "" {
		auth := &authn.Basic{
			Username: username,
			Password: password,
		}
		opts = append(opts, remote.WithAuth(auth))
	}

	tags, err := remote.List(ref, opts...)
	if err != nil {
		log.Fatal(err)
	}

	type tagInfo struct {
		tag     string
		created time.Time
	}
	var tagInfos []tagInfo

	for _, tag := range tags {
		tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", fullRepo, tag))
		if err != nil {
			fmt.Printf("%s: error parsing tag: %v\n", tag, err)
			continue
		}
		img, err := remote.Image(tagRef, opts...)
		if err != nil {
			fmt.Printf("%s: error fetching image: %v\n", tag, err)
			continue
		}
		cfg, err := img.ConfigFile()
		if err != nil {
			fmt.Printf("%s: error fetching config: %v\n", tag, err)
			continue
		}
		created := time.Time{}
		if !cfg.Created.Time.IsZero() {
			created = cfg.Created.Time
		}
		tagInfos = append(tagInfos, tagInfo{tag: tag, created: created})
	}

	// Sort tags by creation date, newest first
	sort.Slice(tagInfos, func(i, j int) bool {
		return tagInfos[i].created.After(tagInfos[j].created)
	})

	// Print all tags with creation date
	for _, ti := range tagInfos {
		created := ""
		if !ti.created.IsZero() {
			created = ti.created.UTC().Format("2006-01-02 15:04:05 MST")
		}
		fmt.Printf("%s\t%s\n", ti.tag, created)
	}

	// Delete older tags if exceeding maxImages
	if maxImages > 0 && len(tagInfos) > maxImages {
		toDelete := tagInfos[maxImages:]
		for _, ti := range toDelete {
			tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", fullRepo, ti.tag))
			if err != nil {
				fmt.Printf("%s: error parsing tag for deletion: %v\n", ti.tag, err)
				continue
			}
			fmt.Printf("Deleting tag: %s\n", ti.tag)
			err = remote.Delete(tagRef, opts...)
			if err != nil {
				fmt.Printf("%s: error deleting tag: %v\n", ti.tag, err)
			}
		}
	}
}
