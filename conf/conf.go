package conf

import (
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

const ConfName = ".review.yml"

type ReviewConf struct {
	Prompt  string   `yaml:"prompt"`
	Exclude []string `yaml:"exclude"`
}

func LoadConf(raw string) (*ReviewConf, error) {
	var cfg ReviewConf
	if err := yaml.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (r *ReviewConf) GetPrompt() string {
	if r == nil {
		return ""
	}

	return r.Prompt
}

func (r *ReviewConf) GetExclude() []string {
	if r == nil {
		return []string{}
	}

	return r.GetExclude()
}

func (r *ReviewConf) MatchAnyPattern(filename string) bool {
	unixPath := filepath.ToSlash(filepath.Clean(filename))

	for _, pattern := range r.GetExclude() {
		if !doublestar.ValidatePattern(pattern) {
			continue
		}

		matched, err := doublestar.Match(pattern, unixPath)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}
