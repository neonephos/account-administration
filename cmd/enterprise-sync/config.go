package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Config is the desired enterprise state, loaded from enterprise/<slug>.yaml.
//
// Only enterprise owners and members are managed. Organizations are
// intentionally out of scope (creating/removing enterprise orgs is destructive
// and, for existing orgs, a manual transfer — see docs).
type Config struct {
	Enterprise string `yaml:"enterprise"`
	Spec       Spec   `yaml:"spec"`
}

type Spec struct {
	// Owners are enterprise administrators (break-glass / foundation admins).
	Owners []string `yaml:"owners"`
	// Members are regular enterprise members.
	Members []string `yaml:"members"`
}

// loadConfig reads and validates a config file.
func loadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true) // reject unknown keys so typos/removed fields fail loudly
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", path, err)
	}
	c.normalize()
	return &c, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.Enterprise) == "" {
		return fmt.Errorf("`enterprise` (slug) is required")
	}
	if len(c.Spec.Owners) == 0 {
		return fmt.Errorf("at least one owner is required")
	}
	// A login cannot be both an owner and a member.
	owners := toSet(c.Spec.Owners)
	for _, m := range c.Spec.Members {
		if owners[strings.ToLower(m)] {
			return fmt.Errorf("%q is listed as both owner and member", m)
		}
	}
	// Reject duplicates within a list.
	if dup := firstDuplicate(c.Spec.Owners); dup != "" {
		return fmt.Errorf("duplicate owner %q", dup)
	}
	if dup := firstDuplicate(c.Spec.Members); dup != "" {
		return fmt.Errorf("duplicate member %q", dup)
	}
	return nil
}

// normalize sorts and de-cases nothing (GitHub logins are case-insensitive but
// we preserve the declared casing); it only sorts for stable diffs.
func (c *Config) normalize() {
	sort.SliceStable(c.Spec.Owners, func(i, j int) bool {
		return strings.ToLower(c.Spec.Owners[i]) < strings.ToLower(c.Spec.Owners[j])
	})
	sort.SliceStable(c.Spec.Members, func(i, j int) bool {
		return strings.ToLower(c.Spec.Members[i]) < strings.ToLower(c.Spec.Members[j])
	})
}

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[strings.ToLower(x)] = true
	}
	return m
}

func firstDuplicate(xs []string) string {
	seen := map[string]bool{}
	for _, x := range xs {
		k := strings.ToLower(x)
		if seen[k] {
			return x
		}
		seen[k] = true
	}
	return ""
}
