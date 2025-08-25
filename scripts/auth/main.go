package main

import (
	"fmt"
	"github.com/htranq/vortech-ome/internal/authorization"
	"github.com/htranq/vortech-ome/pkg/config"
	"time"
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
	signature := auth.Sign(fmt.Sprintf("%s|%s|%s", tableID, serviceID, userID))
	fmt.Println("Signature:", signature)
	fmt.Println("Timestamp:", time.Now().UnixMilli())
}
