package config

import (
	"strings"
)

type DefaultLookup struct {
	AbstractLookup
}

func (e *DefaultLookup) Lookup(key string) string {
	return strings.Split(key, ":")[1]
}

// IsSupport Env lookup support all prefix, so it should be config at the end of the lookup chain
func (e *DefaultLookup) IsSupport(key string) bool {
	return strings.Contains(key, ":")
}

func (e *DefaultLookup) EraseSecret() {

}

func NewDefaultLookup() (Lookupable, error) {
	return &DefaultLookup{}, nil
}
