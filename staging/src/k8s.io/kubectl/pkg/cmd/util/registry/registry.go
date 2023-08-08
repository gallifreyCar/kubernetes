package registry

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"os"
	"path"
	"strings"
)

func getAuth(ref name.Reference) (authn.Authenticator, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fullPath := path.Join(homedir, ".docker/config.json")
	if _, err := os.Stat(fullPath); err == nil {
		return authn.DefaultKeychain.Resolve(ref.Context())
	}
	return nil, nil
}

func GetImageDependenceRaw(image string) (map[string]string, error) {
	ref, err := name.ParseReference(image, name.Insecure)
	if err != nil {
		return nil, err
	}
	auth, err := getAuth(ref)
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(ref, remote.WithAuth(auth))
	if err != nil {
		return nil, err
	}

	images, err := desc.Image()
	if err != nil {
		return nil, err
	}
	cfg, err := images.ConfigFile()
	if err != nil {
		return nil, err
	}

	results := make(map[string]string, len(cfg.Config.Labels))
	for k, v := range cfg.Config.Labels {
		if len(k) <= 4 || !strings.HasPrefix(k, "ver_") {
			continue
		}
		results[k[4:]] = v
	}
	return results, nil
}
