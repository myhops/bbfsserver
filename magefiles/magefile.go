package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = All

func All() {
	mg.Deps(BuildBBFSImageBD)
}

// BuildBBFSImageBD builds a container image and pushes it to docker.io
func BuildBBFSImageBD() error {
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

// BuildBBFSServerLocal build an exe in bin
func BuildBBFSServerLocal() error {
	err := sh.Run("go", "build", "-o" ,"./bin/bbfsserver",  "github.com/myhops/bbfsserver/cmd/bbfsserver")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

func RunBBFSServer() error {
	err := sh.Run("go", "run", "./cmd/bbfsserver/",  "github.com/myhops/bbfsserver/cmd/bbfsserver")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

func RunGodoc() error {
	err := sh.RunV("godoc", "-v", "-http", "localhost:6060", "-index")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}