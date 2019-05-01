package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func runCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
