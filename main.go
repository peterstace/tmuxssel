package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {

	// TODO: Fetch list of existing tmux sessions.

	// TODO: Fetch list of directories we'd wish to create sessions for.

	var sessionToDirMu sync.Mutex
	sessionToDir := make(map[string]string)

	go func() {
		sessions, err := TmuxSessionList()
		if err != nil {
			log.Fatal(err)
		}
		sessionToDirMu.Lock()
		for _, s := range sessions {
			sessionToDir[s] = ""
			// TODO: also send s to fzf
		}
		sessionToDirMu.Unlock()
	}()

	paths := make(chan string)
	go func() {
		if err := filepath.Walk("/home/petsta", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			dir, err := os.Open(path)
			if err != nil {
				return err
			}
			children, err := dir.Readdir(-1)
			if err != nil {
				return err
			}
			for _, child := range children {
				if child.Name() == ".git" {
					paths <- path
					return filepath.SkipDir
				}
			}
			return nil
		}); err != nil {
			log.Fatalf("couldn't walk filesystem: %v", err)
		}
		close(paths)
	}()

	for p := range paths {
		sessionToDirMu.Lock()
		sess := sessionName(p)
		sessionToDir[sess] = p
		// TODO: also send sess to fzf
		sessionToDirMu.Unlock()
	}
}

type Session struct {
	Name string
	Dir  string // empty if session already exists
}

func sessionName(path string) string {
	parts := strings.Split(path, "/")
	return strings.Join(parts[len(parts)-2:], ",")
}
