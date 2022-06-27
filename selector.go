package main

import (
	"log"
	"sync"
)

type selector struct {
	ch             chan string
	sessionToDirMu sync.Mutex
	sessionToDir   map[string]string
	pathToSession  func(string) string
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
	name := s.pathToSession(path)
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
