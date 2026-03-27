package fileMonitor

import (
	"path/filepath"
	"strings"
)

type SuffixStatus int

const (
	SuffixAllowed SuffixStatus = iota
	SuffixSuspicious
	SuffixBlocked
)

type SuffixChecker struct {
	whiteSet map[string]struct{}
	blackSet map[string]struct{}
}

func NewSuffixChecker(whitelist, blacklist []string) *SuffixChecker {
	normalize := func(list []string) map[string]struct{} {
		m := make(map[string]struct{})
		for _, ext := range list {
			ext = strings.ToLower(strings.TrimSpace(ext))
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			m[ext] = struct{}{}
		}
		return m
	}

	return &SuffixChecker{
		whiteSet: normalize(whitelist),
		blackSet: normalize(blacklist),
	}
}

func (sc *SuffixChecker) Check(path string) SuffixStatus {
	ext := strings.ToLower(filepath.Ext(path))
	if _, ok := sc.whiteSet[ext]; ok {
		return SuffixAllowed
	}
	if _, ok := sc.blackSet[ext]; ok {
		return SuffixBlocked
	}
	return SuffixSuspicious
}
