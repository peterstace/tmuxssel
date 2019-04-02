package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	out, err := exec.Command("tmux", "-V").CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			fmt.Println(string(out))
		}
		fmt.Println("running tmux:", err)
		os.Exit(1)
	}
}

func TmuxSessionList() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("tmux: %v %v", err, string(out))
	}
	var sessions []string
	for _, session := range strings.Split(string(out), "\n") {
		session = strings.TrimSpace(session)
		if session != "" {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func TmuxSessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", "="+name)
	err := cmd.Run()
	if err == nil {
		return true
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false
	}
	panic(err)
}

func TmuxNewSession(name, dir string) {
	cmd := exec.Command("tmux", "new-session", "-s", name, "-c", dir, "-d")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func TmuxSwitchSession(name string) {
	subcommand := "switch-client"
	if os.Getenv("TMUX") == "" {
		subcommand = "attach-session"
	}
	cmd := exec.Command("tmux", subcommand, "-t", "="+name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
