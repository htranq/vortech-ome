package main

import (
	"fmt"
	"time"

	"github.com/htranq/vortech-ome/internal/authorization"
	"github.com/htranq/vortech-ome/pkg/config"
)

const _secretKey = "mLwmcZMbhiZfCJiIsMwcslzE55IwmUBG"

func main() {
	auth, err := authorization.New(&config.Authorization{
		Enabled:   true,
		SecretKey: _secretKey,
	})
	if err != nil {
		panic(err)
	}

	var (
		tableID   = "table1"
		serviceID = "service1"
		userID    = "user1"
	)
	timestamp := time.Now().UnixMilli()
	canonical := fmt.Sprintf("%s|%s|%s", tableID, serviceID, userID)
	signature := auth.Sign(canonical)

	fmt.Printf("table_id: %s\n", tableID)
	fmt.Printf("service_id: %s\n", serviceID)
	fmt.Printf("user_id: %s\n", userID)
	fmt.Printf("signature: %s\n", signature)
	fmt.Printf("timestamp: %d\n", timestamp)

	fmt.Println("\ngrpcurl command:")
	fmt.Printf(`grpcurl -plaintext \
  -d '{
    "table_id": "%s",
    "service_id": "%s", 
    "user_id": "%s",
    "authorization": {
      "signature": "%s",
      "timestamp": %d
    }
  }' \
  localhost:8080 \
  vortech.stream_management.management.Management/GetPlaybackUrl
`, tableID, serviceID, userID, signature, timestamp)
}
