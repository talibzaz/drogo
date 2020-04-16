// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	//mg.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "drogo", "./src/code.drogo.gw.com")
	//cmd.Env = []string{"CGO_ENABLED=0"}
	//cmd.Stderr= os.Stderr
	//fmt.Println(cmd.Output())

	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./drogo", "/usr/bin/drogo")
}


// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command( "govendor" ,"sync")
	cmd.Dir = "./src/code.drogo.gw.com"
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("drogo")
}

func GraphGen() error  {
	cmd := exec.Command("gqlgen", "-schema", "./src/code.drogo.gw.com/gql/schema.graphql", "-package", "gql")
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.Rename("./generated.go", "./src/code.drogo.gw.com/gql/generated.go"); err != nil {
		return err
	}
	if err := os.Rename("./models_gen.go", "./src/code.drogo.gw.com/gql/models_gen.go"); err != nil {
		return err
	}
	return nil
}

func MakeDockerImage() {

}