package config

import "strings"

type Lookupable interface {
	Lookup(key string) string
	IsSupport(key string) bool
	EraseSecret() //Clean secret from memory after loaded
}

type AbstractLookup struct {
	Lookupable
}

func (l *AbstractLookup) Key(key string) string {
	return strings.Split(key, ":")[0]
}
