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

func execCmd(args ...string) {
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		panic(string(out))
	}
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
	execCmd("tmux", "new-session", "-s", name, "-c", dir, "-d")
}

func TmuxSwitchSession(name string) {
	if os.Getenv("TMUX") == "" {
		execCmd("tmux", "attach-session", "-t", "="+name)
	} else {
		execCmd("tmux", "switch-client", "-t", "="+name)
	}
}
