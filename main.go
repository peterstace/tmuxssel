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

func home() string {
	if h, ok := os.LookupEnv("HOME"); ok {
		return h
	}
	return "/"
}

type findAndReplace struct {
	find, replace string
}

func main() {
	var ignore, findAndReplaces stringsFlag
	var walkStart string
	flag.Var(&ignore, "ignore", "path fragments to ignore when searching for git repos")
	flag.Var(&findAndReplaces, "find-and-replace", "git-repo-path to session-name find-and-replace directives (in the format FIND:REPLACE)")
	flag.StringVar(&walkStart, "walk-start", home(), "path to start walking from")
	flag.Parse()

	var fnrs []findAndReplace
	for _, pair := range findAndReplaces.values {
		find, replace, ok := strings.Cut(pair, ":")
		if !ok {
			fmt.Fprint(os.Stderr, "rename directive must be in the format: FIND:REPLACE")
			os.Exit(1)
		}
		fnrs = append(fnrs, findAndReplace{find, replace})
	}

	run(walkStart, ignore.values, fnrs)
}

func run(walkStart string, ignore []string, fnrs []findAndReplace) {
	sel := selector{
		ch:           make(chan string, 16),
		sessionToDir: make(map[string]string),
		pathToSession: func(path string) string {
			for _, fnr := range fnrs {
				path = strings.ReplaceAll(path, fnr.find, fnr.replace)
			}
			return path
		},
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
		if err := filepath.Walk(walkStart, func(path string, info os.FileInfo, err error) error {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			if !info.IsDir() {
				return nil
			}
			for _, ig := range ignore {
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
