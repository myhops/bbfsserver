package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = All

func All() {
	mg.Deps(BuildBBFSImageBD)
}

func getGitShortHash() (string, error) {
	return sh.Output("git", "rev-parse", "--short", "HEAD")
}

func gitGetLatestTag() (string, error) {
	return sh.Output("git", "describe", "--tags", "--abbrev=0")
}

// BuildBBFSImageBD builds a container image and pushes it to docker.io
func BuildBBFSImageBD() error {
	env := map[string]string{
		"KO_DOCKER_REPO": "cir-cn.chp.belastingdienst.nl/zandp06",
		"KO_DEFAULTBASEIMAGE": "cir-cn.chp.belastingdienst.nl/zandp06/cgr.dev/chainguard/static:latest-certs",
	}
	err := sh.RunWith(env,
		"ko", "build", "./cmd/bbfsserver")
	if err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	return nil
}

// BuildBBFSImageLocal builds a container image and pushes it to the local docker daemon
func BuildBBFSImageLocal() error {
	// collect the tags
	tags := []string{"latest"}

	t, err := getGitShortHash()
	if err != nil {
		return err
	}
	if t != "" {
		tags = append(tags, t)
	}

	t, err = gitGetLatestTag()
	if err != nil {
		return err
	}
	if t != "" {
		tags = append(tags, t)
	}

	imageTags := strings.Join(tags, ",")

	err = sh.Run("ko", "build", "--local", "--tags", imageTags, "./cmd/bbfsserver")
	if err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	return nil
}

// BuildBBFSServerLocal build an exe in bin
func BuildBBFSServerLocal() error {
	err := sh.Run("go", "build", "-o", "./bin/bbfsserver", "github.com/myhops/bbfsserver/cmd/bbfsserver")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

func RunBBFSServer() error {
	err := sh.Run("go", "run", "./cmd/bbfsserver/", "github.com/myhops/bbfsserver/cmd/bbfsserver")
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
