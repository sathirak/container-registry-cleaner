package main

import (
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sethvargo/go-githubactions"
)

func main() {

	repoName := githubactions.GetInput("name")
	if repoName == "" {
		githubactions.Fatalf("missing 'name'")
	}

	ref, err := name.NewRepository(repoName)
	if err != nil {
        log.Fatal(err)
    }

    tags, err := remote.List(ref)
    if err != nil {
        log.Fatal(err)
    }

    for _, tag := range tags {
        fmt.Println(tag)
    }

	fmt.Println("Hello, world!")
}
