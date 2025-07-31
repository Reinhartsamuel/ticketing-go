package config

import (
	"log"

	"github.com/joho/godotenv"
)

func Get() *Config {
	err := godotenv.Load()
	if err != nil {
		// panic(err)
		log.Fatal("Error loading environment variables::::", err.Error())
	}
	return &Config{
		// Server : Server{
		// 	Host: os.Getenv("SERVER_HOST"),
		// 	Port: os.Getenv("SERVER_PORT"),
		// },
		// Database: Database{
		// 	Host: os.Getenv("DB_HOST"),
		// 	Port: os.Getenv("DB_PORT"),
		// 	Name: os.Getenv("DB_NAME"),
		// 	User: os.Getenv("DB_USER"),
		// 	Pass: os.Getenv("DB_PASS"),
		// 	Tz:   os.Getenv("DB_TZ"),
		// }
		Server: Server{
			Host: "test",
			Port: "8080",
		},
		Database: Database{
			Host: "localhost",
			Port: "5432",
			Name: "testdb",
			User: "user",
			Pass: "pass",
			Tz:   "UTC",
		},
	}
}
