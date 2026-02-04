package main

import (
	"auren-platform/config"
	"auren-platform/db"
	"log"
)

func main() {
	cfg := config.Load()
	_ = cfg

	// Placeholder server entry that will be expanded in the refactor.
	// Currently the project's root `main.go` is still the runtime entrypoint.
	log.Println("cmd/server placeholder: use root main.go during migration. Config loaded.")
	db.InitDB()
}
