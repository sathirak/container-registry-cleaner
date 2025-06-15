package cr

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sathirak/container-registry-cleaner/internal/registry"
	"github.com/sathirak/container-registry-cleaner/internal/types"
)

func CleanRegistry(config types.Config) error {
	ref, err := name.NewRepository(config.RepoName)
	if err != nil {
		return err
	}

	var opts []remote.Option
	if config.Registry == "docr" || config.Registry == "registry.digitalocean.com" {
		opts, err = registry.DOCRAuthOptions(config.Password)
		if err != nil {
			return err
		}
	} else if config.Username != "" && config.Password != "" {
		auth := &authn.Basic{
			Username: config.Username,
			Password: config.Password,
		}
		opts = append(opts, remote.WithAuth(auth))
	}

	tags, err := remote.List(ref, opts...)
	if err != nil {
		return err
	}

	var tagInfos []types.TagInfo

	for _, tag := range tags {
		tagRef, err := name.NewTag(fmt.Sprintf("%s:%s", config.RepoName, tag))
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
		tagInfos = append(tagInfos, types.TagInfo{Tag: tag, Created: created})
	}

	sort.Slice(tagInfos, func(i, j int) bool {
		return tagInfos[i].Created.After(tagInfos[j].Created)
	})

	fmt.Printf("%-30s %-25s\n", "Tag", "Created At")
	fmt.Printf("%s\n", "---------------------------------------------------------------")
	for _, ti := range tagInfos {
		created := ""
		if !ti.Created.IsZero() {
			created = ti.Created.UTC().Format("2006-01-02 15:04:05 MST")
		}
		fmt.Printf("%-30s %-25s\n", ti.Tag, created)
	}

	return nil
}
