package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func init() {
	// Check to make sure fzf is installed.
	if err := exec.Command("fzf", "--version").Run(); err != nil {
		log.Fatal("%v", err)
	}
}

func FZF(input <-chan string) (string, bool, error) {
	cmd := exec.Command("fzf")
	cmd.Stdout = new(bytes.Buffer)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", false, err
	}
	go func() {
		defer stdin.Close()
		for in := range input {
			fmt.Fprintf(stdin, "%s\n", in)
		}
	}()

	if err := cmd.Run(); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// Ctrl-C or bad input provided.
			return "", false, nil
		default:
			return "", false, err
		}
	}
	return string(bytes.TrimRight(cmd.Stdout.(*bytes.Buffer).Bytes(), "\n")), true, nil
}
