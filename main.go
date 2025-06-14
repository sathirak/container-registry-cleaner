package main

import (
	"fmt"
	"log"

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
		created := ""
		if !cfg.Created.Time.IsZero() {
			created = cfg.Created.UTC().Format("2006-01-02 15:04:05 MST")
		}
		fmt.Printf("%s\t%s\n", tag, created)
	}
}
