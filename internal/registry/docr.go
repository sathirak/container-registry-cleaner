package registry

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func DOCRAuthOptions(password string) ([]remote.Option, error) {
	if password == "" {
		return nil, fmt.Errorf("missing DigitalOcean API token for DOCR")
	}
	auth := &authn.Basic{
		Username: "doctoken",
		Password: password,
	}
	return []remote.Option{remote.WithAuth(auth)}, nil
}
