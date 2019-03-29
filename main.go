package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	var sessionToDirMu sync.Mutex
	sessionToDir := make(map[string]string)

	fzfCh := make(chan string)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		sessions, err := TmuxSessionList()
		if err != nil {
			log.Fatal(err)
		}
		sessionToDirMu.Lock()
		for _, s := range sessions {
			sessionToDir[s] = ""
		}
		sessionToDirMu.Unlock()
		for _, s := range sessions {
			fzfCh <- s
		}
		wg.Done()
	}()

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
					sessionToDirMu.Lock()
					sess := sessionName(path)
					sessionToDir[sess] = path
					sessionToDirMu.Unlock()
					fzfCh <- sess
					return filepath.SkipDir
				}
			}
			return nil
		}); err != nil {
			log.Fatalf("couldn't walk filesystem: %v", err)
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(fzfCh)
	}()

	got, ok, err := FZF(fzfCh)
	if err != nil {
		log.Fatalf("fzf: %v", err)
	}

	if !TmuxSessionExists(got) {
		TmuxNewSession(got, sessionToDir[got])
	}
	TmuxSwitchSession(got)
}

func sessionName(path string) string {
	parts := strings.Split(path, "/")
	sess := strings.Join(parts[len(parts)-2:], "/")
	return strings.ReplaceAll(sess, ".", ",")
}
