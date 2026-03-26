package container

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
)

func NewClient() (*docker.Client, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("connecting to Docker: %w\nIs Docker running?", err)
	}
	// Negotiate API version with the Docker daemon so the client works
	// with any Docker Engine version (old library default is too low for
	// modern Docker Desktop).
	if env, err := cli.Version(); err == nil {
		if apiVer := env.Get("ApiVersion"); apiVer != "" {
			if negotiated, err := docker.NewVersionedClientFromEnv(apiVer); err == nil {
				return negotiated, nil
			}
		}
	}
	return cli, nil
}
