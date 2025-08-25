package authorization

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/htranq/vortech-ome/pkg/config"
)

func TestAuthorization(t *testing.T) {
	cfg := &config.Authorization{
		Enabled:   true,
		SecretKey: "test-secret-key",
	}

	auth, err := New(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, auth)

	canonical := fmt.Sprintf("%d|%s|%s|%s", time.Now().Second(), "table1", "service1", "user1")

	// Sign
	signature := auth.Sign(canonical)
	assert.NotEmpty(t, signature)

	// Verify - success
	err = auth.Verify(canonical, signature)
	assert.NoError(t, err)

	// Verify - failure
	err = auth.Verify(canonical, "invalid-signature")
	assert.Error(t, err)
}
