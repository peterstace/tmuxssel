package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type stringsFlag struct {
	values []string
}

func (s *stringsFlag) String() string {
	if s == nil {
		return "[]"
	}
	return fmt.Sprintf("%v", s.values)
}

func (s *stringsFlag) Set(v string) error {
	s.values = append(s.values, v)
	return nil
}

func main() {
	var ignore stringsFlag
	flag.Var(&ignore, "ignore", "path fragments to ignore when searching for git repos")
	flag.Parse()

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
			for _, ig := range ignore.values {
				if strings.Contains(path, ig) {
					return filepath.SkipDir
				}
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
