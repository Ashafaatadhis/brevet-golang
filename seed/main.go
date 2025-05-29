package main

import (
	"brevet-api/config"
	"brevet-api/seed/master"
	"fmt"
	"log"
)

func main() {
	db := config.ConnectDB()

	fmt.Println("Seeding roles...")
	if err := master.SeedRoles(db); err != nil {
		log.Fatalf("failed seeding roles: %v", err)
	}

	fmt.Println("Seeding course types...")
	if err := master.SeedCourseTypes(db); err != nil {
		log.Fatalf("failed seeding course types: %v", err)
	}

	fmt.Println("Seeding groups...")
	if err := master.SeedGroups(db); err != nil {
		log.Fatalf("failed seeding groups: %v", err)
	}

	fmt.Println("Seeding prices...")
	if err := master.SeedPrices(db); err != nil {
		log.Fatalf("failed seeding prices: %v", err)
	}

	fmt.Println("Seeding days...")
	if err := master.SeedDays(db); err != nil {
		log.Fatalf("failed seeding days: %v", err)
	}

	fmt.Println("Seeding done!")
}
