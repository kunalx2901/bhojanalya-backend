package main

import (
	"log"

	"bhojanalya/internal/auth"
	"bhojanalya/internal/router"
)

func main() {
	repo := auth.NewInMemoryUserRepository()
	service := auth.NewService(repo)

	r := router.NewRouter(service)

	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}
