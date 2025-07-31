package main

import (
	"ticketing-go/internal/api"
	"ticketing-go/internal/config"
	"ticketing-go/internal/connection"
	"ticketing-go/internal/repository"
	"ticketing-go/internal/services"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cnf := config.Get()
	dbConnection := connection.GetDatabase(cnf.Database)

	app := fiber.New()

	customerRepository := repository.NewCustomer(dbConnection)
	customerService := services.NewCustomer(customerRepository)

	api.NewCustomer(app, customerService)

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.JSON("Hello, World!")
	})
	app.Get("/developers", func(ctx *fiber.Ctx) error {
		return ctx.JSON("Hello, developers!")
	})
	_ = app.Listen(cnf.Server.Host + ":" + cnf.Server.Port)
}

func developers(ctx *fiber.Ctx) error {
	return ctx.SendString("Hello, World!")
}
