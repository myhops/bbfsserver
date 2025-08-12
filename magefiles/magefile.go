package main

import (
	"fmt"
	"os"
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

func getAllTags() (string, error) {
	// collect the tags
	tags := []string{"latest"}

	t, err := getGitShortHash()
	if err != nil {
		return "", err
	}
	if t != "" {
		tags = append(tags, t)
	}

	t, err = gitGetLatestTag()
	if err != nil {
		return "", err
	}
	if t != "" {
		tags = append(tags, t)
	}

	return strings.Join(tags, ","), nil
}

// BuildBBFSImageBD builds a container image and pushes it to cir-cn.chp.belastingdienst.nl/zandp06
func BuildBBFSImageBD() error {
	env := map[string]string{
		// "KO_DOCKER_REPO":      "cir-cn-devops.chp.belastingdienst.nl/obp-pnr",
		"KO_DOCKER_REPO":      "ko.local",
		"KO_DEFAULTBASEIMAGE": "cir-cn-cpet.chp.belastingdienst.nl/external/docker.io/alpine:3.22.1",
		"DOCKER_HOST":         "unix:///tmp/podman.sock",
	}

	// podman pull cir-cn-devops.chp.belastingdienst.nl/obp-pnr/bbfsserver-313e8234ab7c35c20f5af54de96e0417:latest
	// podman pull cir-cn-cpet.chp.belastingdienst.nl/external/docker.io/alpine:3.22.1

	imageTags, err := getAllTags()
	if err != nil {
		return err
	}

	if err := sh.RunWith(env, "ko", "build", "--tags", imageTags, "./cmd/bbfsserver"); err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	return nil
}

var BDImages = []string{
	"cir-cn-devops.chp.belastingdienst.nl/obp-pnr/bbfsserver:latest",
	"cir-cn-devops.chp.belastingdienst.nl/obp-pnr/bbfsserver:v0.0.12",
}

func PublishFromDockerfile() error {
	mg.Deps(BuildDockerfile)

	// Push the first image
	{
		image := BDImages[0]
		if err := sh.Run("podman", "push", image); err != nil {
			return fmt.Errorf("error pusing %s: %w", image, err)
		}
	}

	for _, image := range BDImages[1:] {
		// split the image and the tag
		parts := strings.Split(image, ":")
		if len(parts) != 2 {
			continue
		}
		if err := sh.Run("crane", "tag", parts[0], parts[1]); err != nil {
			return fmt.Errorf("error tagging %s with %s: %w", parts[0], parts[1], err)
		}
	}
	return nil
}

func CopyCerts() error {
	os.MkdirAll("certs", 0755)
	return sh.Run("cp", "/etc/ssl/certs/ca-certificates.crt", "certs")
}

func BuildDockerfile() error {
	mg.Deps(CopyCerts)

	args := []string{"build"}
	for _, i := range BDImages {
		args = append(args, "--tag", i)
	}
	args = append(args, ".")

	if err := sh.Run("podman", args...); err != nil {
		return fmt.Errorf("error building image: %w", err)
	}
	return nil
}

// BuildBBFSImageLocal builds a container image and pushes it to the local docker daemon
func BuildBBFSImageLocal() error {
	imageTags, err := getAllTags()
	if err != nil {
		return err
	}

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
