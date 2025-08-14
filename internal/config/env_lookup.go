package config

import (
	"fmt"
	"os"
)

type EnvLookup struct {
	AbstractLookup
}

func (e *EnvLookup) Lookup(key string) string {
	k := e.Key(key)
	if val, found := os.LookupEnv(k); found {
		return val
	}
	return fmt.Sprintf("${%s}", key)

}

// IsSupport Env lookup support all prefix, so it should be config at the end of the lookup chain
func (e *EnvLookup) IsSupport(key string) bool {
	_, exist := os.LookupEnv(e.Key(key))
	return exist
}

func (e *EnvLookup) EraseSecret() {

}

func NewEnvLookup() (Lookupable, error) {
	return &EnvLookup{}, nil
}
