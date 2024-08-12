package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = All

func All() {
	mg.Deps(BuildBBFSServer)
}

// BuildKedaplay builds a container image and pushes it to docker.io
func BuildBBFSServer() error {
	env := map[string]string{
		"KO_DOCKER_REPO": "cir-cn.chp.belastingdienst.nl/zandp06",
	}
	err := sh.RunWith(env,
		"ko", "build", "./cmd/bbfsserver")
	if err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	return nil
}

