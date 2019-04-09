package main

import (
	"log"
	"strings"
	"sync"
)

type selector struct {
	ch             chan string
	sessionToDirMu sync.Mutex
	sessionToDir   map[string]string
}

func (s *selector) fzfSelect() (string, string, bool) {
	got, ok, err := FZF(s.ch)
	if err != nil {
		log.Fatalf("fzf: %v", err)
	}
	return got, s.sessionToDir[got], ok
}

func (s *selector) addExistingSession(name string) {
	s.sessionToDirMu.Lock()
	if _, ok := s.sessionToDir[name]; !ok {
		s.ch <- name
		s.sessionToDir[name] = ""
	}
	s.sessionToDirMu.Unlock()
}

func (s *selector) addPath(path string) {
	name := sessionName(path)
	s.sessionToDirMu.Lock()
	if _, ok := s.sessionToDir[name]; !ok {
		s.ch <- name
		s.sessionToDir[name] = path
	}
	s.sessionToDirMu.Unlock()
}

func (s *selector) finishedAdding() {
	close(s.ch)
}

func sessionName(path string) string {
	parts := strings.Split(path, "/")
	sess := strings.Join(parts[len(parts)-2:], "/")
	return strings.ReplaceAll(sess, ".", ",")
}
