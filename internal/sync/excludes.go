package sync

import (
	"path/filepath"
	"strings"
)

type ExcludeRules struct {
	patterns []string
}

func NewExcludeRules(patterns []string) *ExcludeRules {
	return &ExcludeRules{
		patterns: patterns,
	}
}

func (r *ExcludeRules) ShouldExclude(path string) bool {
	for _, pattern := range r.patterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}

		if strings.HasSuffix(pattern, "/") {
			dirName := strings.TrimSuffix(pattern, "/")
			if strings.Contains(path, "/"+dirName+"/") || strings.HasSuffix(path, "/"+dirName) {
				return true
			}
		}
	}
	return false
}

func (r *ExcludeRules) AddPattern(pattern string) {
	r.patterns = append(r.patterns, pattern)
}

func (r *ExcludeRules) RemovePattern(pattern string) {
	for i, p := range r.patterns {
		if p == pattern {
			r.patterns = append(r.patterns[:i], r.patterns[i+1:]...)
			return
		}
	}
}
