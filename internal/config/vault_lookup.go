package config

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/hashicorp/vault/api/auth/kubernetes"
)

type VaultParameters struct {
	Enabled    bool   `envconfig:"VAULT_ENABLED" default:"false"`
	Address    string `envconfig:"VAULT_ADDRESS"`
	AuthMethod string `envconfig:"VAULT_AUTH_METHOD"`

	Token string `envconfig:"VAULT_TOKEN"`

	ApproleRoleID       string `envconfig:"VAULT_APPROLE_ROLE_ID"`
	ApproleSecretIDFile string `envconfig:"VAULT_APPROLE_SECRET_ID_FILE"`

	KubernetesRole                    string `envconfig:"VAULT_KUBERNETES_ROLE"`
	KubernetesServiceAccountTokenPath string `envconfig:"VAULT_KUBERNETES_SERVICE_ACCOUNT_TOKEN_PATH"`

	SecretPath string `envconfig:"VAULT_SECRET_PATH"`
}

type VaultLookup struct {
	AbstractLookup
	client     *vault.Client
	parameters VaultParameters
	secret     *vault.Secret
	secretData map[string]interface{}
	loaded     bool
}

func (v *VaultLookup) EraseSecret() {
	v.secret = nil
	v.loaded = false
}

func (v *VaultLookup) IsSupport(originalKey string) bool {
	// Support all key
	key := v.Key(strings.ToUpper(originalKey))
	_, ok := v.secretData[key]
	return ok
}

func (v *VaultLookup) loadSecret() error {
	if v.client == nil {
		return fmt.Errorf("vault client is not initialized")
	}

	secretPath := v.parameters.SecretPath

	if !strings.HasPrefix(secretPath, "kv/data") {
		secretPath = fmt.Sprintf("kv/data/%s", secretPath)
	}

	secret, err := v.client.Logical().Read(secretPath)
	if err != nil {
		return fmt.Errorf("failed to read secret from vault: %s", err)
	}

	if secret == nil {
		return fmt.Errorf("secret not found: %s", secretPath)
	}

	if secret.Data == nil {
		return fmt.Errorf("secret data not found: %s", secretPath)
	}

	data := secret.Data["data"].(map[string]interface{})
	// Convert all key to lower case
	v.secretData = uppercaseMapKey(data)
	v.secret = secret
	v.loaded = true
	return nil
}

func (v *VaultLookup) Lookup(originalKey string) string {
	key := v.Key(strings.ToUpper(originalKey))
	var secret = v.secretData

	value, ok := secret[key]
	if !ok {
		return fmt.Sprintf("${%s}", originalKey) //return original key
	}

	return value.(string)
}

func NewVaultLookup(ctx context.Context, parameters VaultParameters) (Lookupable, error) {
	config := vault.DefaultConfig()
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	v := &VaultLookup{
		client:     client,
		parameters: parameters,
	}
	_, err = v.login(ctx)
	if err != nil {
		return nil, fmt.Errorf("vault login error: %w", err)
	}
	err = v.loadSecret()
	if err != nil {
		return nil, fmt.Errorf("vault load secret error: %w", err)
	}
	return v, nil
}

func (v *VaultLookup) login(ctx context.Context) (*vault.Secret, error) {

	var authMethodParams vault.AuthMethod
	var err error

	authMethod := strings.ToLower(v.parameters.AuthMethod)
	switch authMethod {
	case "approle":
		approleSecretID := &approle.SecretID{
			FromFile: v.parameters.ApproleSecretIDFile,
		}
		authMethodParams, err = approle.NewAppRoleAuth(
			v.parameters.ApproleRoleID,
			approleSecretID,
			approle.WithWrappingToken(), // only required if the SecretID is response-wrapped
		)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize approle authentication method: %w", err)
		}

	case "kubernetes":
		var options []kubernetes.LoginOption
		if v.parameters.KubernetesServiceAccountTokenPath != "" {
			options = append(options, kubernetes.WithServiceAccountTokenPath(v.parameters.KubernetesServiceAccountTokenPath))
		}
		authMethodParams, err = kubernetes.NewKubernetesAuth(
			v.parameters.KubernetesRole, options...,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize kubernetes authentication method: %w", err)
		}
	case "Token":
		v.client.SetToken(v.parameters.Token)

		return nil, nil

	}

	authInfo, err := v.client.Auth().Login(ctx, authMethodParams)
	if err != nil {
		return nil, fmt.Errorf("unable to login using approle auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}
