package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AwsSecretManagerParameter struct {
	Enabled            bool   `envconfig:"AWS_SECRET_ENABLED" default:"false"`
	Region             string `envconfig:"AWS_SECRET_REGION" default:"ap-northeast-1"`
	SecretArn          string `envconfig:"AWS_SECRET_ARN"`
	SecretPath         string `envconfig:"AWS_SECRET_PATH"` // path to field in json value, the field must be a map
	SecretVersion      string `envconfig:"AWS_SECRET_VERSION" `
	SecretVersionStage string `envconfig:"AWS_SECRET_VERSION_STAGE" default:"AWSCURRENT"`
}

type AwsSecretManagerLookup struct {
	AbstractLookup
	data map[string]interface{}
}

func (a *AwsSecretManagerLookup) Lookup(originalKey string) string {
	key := a.Key(strings.ToUpper(originalKey))
	if val, found := a.data[key]; found {
		return val.(string)
	}
	return fmt.Sprintf("${%s}", originalKey)
}

func (a *AwsSecretManagerLookup) IsSupport(originalKey string) bool {
	// Support all key
	key := a.Key(strings.ToUpper(originalKey))
	_, found := a.data[key]
	return found
}

func (a *AwsSecretManagerLookup) EraseSecret() {
	a.data = nil
}

func NullableString(val string) *string {
	if val == "" {
		return nil
	}
	return aws.String(val)
}

func NewAwsSecretManagerLookup(ctx context.Context, parameter AwsSecretManagerParameter) (Lookupable, error) {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	awsConfig.Region = parameter.Region
	if err != nil {
		return nil, err
	}
	client := secretsmanager.NewFromConfig(awsConfig)

	val, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(parameter.SecretArn),
		VersionId:    NullableString(parameter.SecretVersion),
		VersionStage: NullableString(parameter.SecretVersionStage),
	})

	if err != nil {
		return nil, err
	}
	var secretString string
	if val.SecretString != nil {
		secretString = *val.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(val.SecretBinary)))
		l, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, val.SecretBinary)
		if err != nil {
			return nil, err
		}
		secretString = string(decodedBinarySecretBytes[:l])
	}
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	err = json.Unmarshal([]byte(secretString), &result)
	if err != nil {
		return nil, err
	}
	data, err := getMapField(result, parameter.SecretPath)
	if err != nil {
		return nil, err
	}
	data = uppercaseMapKey(data)
	return &AwsSecretManagerLookup{data: data}, nil
}
