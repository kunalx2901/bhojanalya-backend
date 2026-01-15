package main

import (
	"log"
	"bhojanalya/internal/auth"
	"bhojanalya/internal/router"
	"bhojanalya/prisma/db" // Adjust path accordingly
)

func main() {
	// 1. Initialize Prisma Client
	client := db.NewClient()
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Prisma connect error: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			panic(err)
		}
	}()

	// 2. Use Prisma Repository instead of InMemory
	repo := auth.NewPrismaUserRepository(client)
	service := auth.NewService(repo)

	// 3. Start Router
	r := router.NewRouter(service)

	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}