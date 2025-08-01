package main

import (
	"Repos/ticketing-go/helpers"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variables:", err.Error())
	}
	log.Println("Starting USDT listener...")
	helpers.StartListener()
}
