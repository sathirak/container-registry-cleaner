package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sethvargo/go-githubactions"
)

func deleteDigitalOceanTag(token, repo, tag string) error {
	// repo is expected to be like "namespace/repo"
	url := fmt.Sprintf("https://api.digitalocean.com/v2/registry/repositories/%s/tags/%s", repo, tag)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

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

	var fullRepo string
	switch registry {
	case "ghcr":
		fullRepo = fmt.Sprintf("ghcr.io/%s", repoName)
	case "dockerhub":
		if strings.Contains(repoName, "/") {
			fullRepo = repoName
		} else {
			fullRepo = fmt.Sprintf("library/%s", repoName)
		}
	case "digitalocean":
		fullRepo = fmt.Sprintf("registry.digitalocean.com/%s", repoName)
	default:
		githubactions.Fatalf("unsupported registry: %s", registry)
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
			// Skip deletion for tags that look like digests (e.g., sha-xxxxxx)
			if len(ti.tag) > 4 && ti.tag[:4] == "sha-" {
				fmt.Printf("Skipping deletion for digest-like tag: %s (unsupported by registry)\n", ti.tag)
				continue
			}

			switch registry {
			case "digitalocean":
				fmt.Printf("Deleting tag from DigitalOcean: %s\n", ti.tag)
				err := deleteDigitalOceanTag(password, repoName, ti.tag)
				if err != nil {
					fmt.Printf("%s: error deleting tag from DigitalOcean: %v\n", ti.tag, err)
				}
			case "ghcr":
				tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", fullRepo, ti.tag))
				if err != nil {
					fmt.Printf("%s: error parsing tag for deletion: %v\n", ti.tag, err)
					continue
				}
				fmt.Printf("Deleting tag from GHCR: %s\n", ti.tag)
				err = remote.Delete(tagRef, opts...)
				if err != nil {
					fmt.Printf("%s: error deleting tag from GHCR: %v\n", ti.tag, err)
				}
			case "dockerhub":
				tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", fullRepo, ti.tag))
				if err != nil {
					fmt.Printf("%s: error parsing tag for deletion: %v\n", ti.tag, err)
					continue
				}
				fmt.Printf("Deleting tag from Docker Hub: %s\n", ti.tag)
				err = remote.Delete(tagRef, opts...)
				if err != nil {
					fmt.Printf("%s: error deleting tag from Docker Hub: %v\n", ti.tag, err)
				}
			}
		}
	}
}
