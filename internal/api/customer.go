package api

import (
	"context"
	"net/http"
	"ticketing-go/domain"

	"github.com/gofiber/fiber/v2"

	"ticketing-go/dto"
	"time"
)

type customerApi struct {
	customerService domain.CustomerService
}

func NewCustomer(app *fiber.App, customerService domain.CustomerService) {
	ca := customerApi{
		customerService: customerService,
	}
	app.Get("/customers", ca.Index)
	app.Get("/customers/:id", ca.FindByID)
}

func (ca customerApi) Index(ctx *fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	res, err := ca.customerService.Index(c)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).
			JSON(dto.CreateResponseError(err.Error()))
	}
	// return ctx.Status(http.StatusOK).
	// JSON(dto.CreateREsponseSuccess(res))
	return ctx.Status(http.StatusOK).
		JSON(dto.CreateResponseSuccess(res))
}
func (ca customerApi) FindByID(ctx *fiber.Ctx) error {
	c, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	res, err := ca.customerService.FindByID(c, ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).
			JSON(dto.CreateResponseError(err.Error()))
	}
	// return ctx.Status(http.StatusOK).
	// JSON(dto.CreateREsponseSuccess(res))
	return ctx.Status(http.StatusOK).
		JSON(dto.CreateResponseSuccess(res))
}
