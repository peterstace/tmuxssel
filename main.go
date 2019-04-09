package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	//var sessionToDirMu sync.Mutex
	//sessionToDir := make(map[string]string)

	//fzfCh := make(chan string)

	sel := selector{
		ch:           make(chan string, 16),
		sessionToDir: make(map[string]string),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		sessions, err := TmuxSessionList()
		if err != nil {
			log.Fatal(err)
		}
		for _, s := range sessions {
			sel.addExistingSession(s)
		}
		wg.Done()
	}()

	go func() {
		if err := filepath.Walk(home(), func(path string, info os.FileInfo, err error) error {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			if !info.IsDir() {
				return nil
			}
			if strings.Contains(path, ".cargo/registry/index") {
				return filepath.SkipDir
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
					sel.addPath(path)
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
		sel.finishedAdding()
	}()

	sess, dir, ok := sel.fzfSelect()
	if !ok {
		return
	}

	if !TmuxSessionExists(sess) {
		TmuxNewSession(sess, dir)
	}
	TmuxSwitchSession(sess)
}

func home() string {
	if h, ok := os.LookupEnv("HOME"); ok {
		return h
	}
	return "/"
}
