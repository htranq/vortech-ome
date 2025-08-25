// Management gRPC Client Script
//
// This script makes a gRPC call to GetPlaybackUrl service running on localhost.
// It generates proper authorization signatures and returns a WebSocket URL with stream token.
//
// Usage:
//
//	go run scripts/management/main.go
//	go run scripts/management/main.go -table "baccarat_table" -user "player123"
//	go run scripts/management/main.go -addr "localhost:9090" -service "live_stream"
//
// Flags:
//
//	-addr    gRPC server address (default: localhost:8080)
//	-table   Table ID (default: table1)
//	-service Service ID (default: service1)
//	-user    User ID (default: user1)
//	-secret  Secret key for authorization
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	managementpb "github.com/htranq/vortech-ome/api/v1/management"
	"github.com/htranq/vortech-ome/internal/authorization"
	"github.com/htranq/vortech-ome/pkg/config"
)

const _secretKey = "mLwmcZMbhiZfCJiIsMwcslzE55IwmUBG"

func main() {
	// Command-line flags
	var (
		tableID    = flag.String("table", "table1", "Table ID")
		serviceID  = flag.String("service", "service1", "Service ID")
		userID     = flag.String("user", "user1", "User ID")
		serverAddr = flag.String("addr", "localhost:8080", "gRPC server address")
		secretKey  = flag.String("secret", _secretKey, "Secret key for authorization")
	)
	flag.Parse()

	auth, err := authorization.New(&config.Authorization{
		Enabled:   true,
		SecretKey: *secretKey,
	})
	if err != nil {
		log.Fatalf("Failed to create authorization: %v", err)
	}

	// Create gRPC connection
	fmt.Printf("🔗 Connecting to gRPC server at %s...\n", *serverAddr)
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Create management client
	client := managementpb.NewManagementClient(conn)

	// Prepare request parameters
	timestamp := time.Now().UnixMilli()
	canonical := fmt.Sprintf("%s|%s|%s", *tableID, *serviceID, *userID)
	signature := auth.Sign(canonical)

	fmt.Printf("📋 Request parameters:\n")
	fmt.Printf("  table_id: %s\n", *tableID)
	fmt.Printf("  service_id: %s\n", *serviceID)
	fmt.Printf("  user_id: %s\n", *userID)
	fmt.Printf("  signature: %s\n", signature)
	fmt.Printf("  timestamp: %d\n", timestamp)
	fmt.Printf("  canonical: %s\n", canonical)

	// Create request
	request := &managementpb.GetPlaybackUrlRequest{
		TableId:   *tableID,
		ServiceId: *serviceID,
		UserId:    *userID,
		Authorization: &managementpb.Authorization{
			Signature: signature,
			Timestamp: timestamp,
		},
		// Optional: set expiration time (uncomment if needed)
		// ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
	}

	// Make the gRPC call
	fmt.Printf("\n🚀 Making gRPC call to GetPlaybackUrl...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.GetPlaybackUrl(ctx, request)
	if err != nil {
		log.Fatalf("❌ GetPlaybackUrl failed: %v", err)
	}

	// Display response
	fmt.Printf("\n✅ Success! Response received:\n")
	fmt.Printf("  url: %s\n", response.GetUrl())
	fmt.Printf("  stream_token: %s\n", response.GetStreamToken())
}
