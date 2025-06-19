package main

import (
	"brevet-api/config"
	"brevet-api/seed/master"
	"fmt"
	"log"
)

func main() {
	db := config.ConnectDB()

	fmt.Println("Seeding prices...")
	if err := master.SeedPrices(db); err != nil {
		log.Fatalf("failed seeding prices: %v", err)
	}

	fmt.Println("Seeding done!")
}
