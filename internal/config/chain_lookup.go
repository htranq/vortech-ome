package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

type ChainLookup struct {
	chain  []Lookupable
	logger *zap.Logger
}

func (l *ChainLookup) EraseSecret() {
	for _, looker := range l.chain {
		looker.EraseSecret()
	}
}

func (l *ChainLookup) IsSupport(key string) bool {
	for _, looker := range l.chain {
		if looker.IsSupport(key) {
			return true
		}
	}
	return false
}

func (l *ChainLookup) Lookup(key string) string {
	for _, looker := range l.chain {
		if looker.IsSupport(key) {
			ret := looker.Lookup(key)
			return strings.ReplaceAll(ret, "\n", "\\n")
		}
	}
	l.logger.Warn(fmt.Sprintf("key %s not found in any lookup chain", key))
	return fmt.Sprintf("${%s}", key) //Return original key if not exist
}

func NewLookupChain(chain ...Lookupable) (Lookupable, error) {
	c := zap.NewProductionConfig()
	c.DisableStacktrace = true
	l, err := c.Build()
	if err != nil {
		panic(err)
	}

	return &ChainLookup{chain: chain, logger: l}, nil
}

func NewLookupChainFromEnv() (Lookupable, error) {
	var chain []Lookupable
	var ctx = context.Background()

	// Vault provider
	vaultParams := VaultParameters{}
	err := envconfig.InitWithOptions(&vaultParams, envconfig.Options{
		Prefix:          "VAULT",
		LeaveNil:        true,
		AllOptional:     true,
		AllowUnexported: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init vault parameters: %s", err)
	}

	if vaultParams.Enabled {
		vault, err := NewVaultLookup(ctx, vaultParams)
		if err != nil {
			return nil, fmt.Errorf("failed to init vault lookup: %s", err)
		}
		chain = append(chain, vault)
	}

	//Aws secret manager provider
	awsParams := AwsSecretManagerParameter{}
	err = envconfig.InitWithOptions(&awsParams, envconfig.Options{
		Prefix:          "AWS_SECRET",
		LeaveNil:        true,
		AllOptional:     true,
		AllowUnexported: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init aws secret manager parameters: %s", err)
	}

	if awsParams.Enabled {
		aws, err := NewAwsSecretManagerLookup(ctx, awsParams)
		if err != nil {
			return nil, fmt.Errorf("failed to init aws secret manager lookup: %s", err)
		}
		chain = append(chain, aws)
	}
	// Env provider

	envLookup, err := NewEnvLookup()
	if err != nil {
		return nil, fmt.Errorf("failed to init env lookup: %s", err)
	}
	chain = append(chain, envLookup)

	// More providers can be added here
	defaultLookup, err := NewDefaultLookup()
	if err != nil {
		return nil, fmt.Errorf("failed to init default lookup: %s", err)
	}
	chain = append(chain, defaultLookup)
	return NewLookupChain(chain...)
}
