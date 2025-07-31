package domain

import (
	"context"
	"database/sql"
	"ticketing-go/dto"
)

type Customer struct {
	ID        string       `db:"id"`
	Code      string       `db:"code"`
	Name      string       `db:"name"`
	Email     string       `db:"email"`
	Address   string       `db:"address"`
	Phone     string       `db:"phone"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type CustomerRepository interface {
	FindAll(ctx context.Context) ([]Customer, error)
	FindByID(ctx context.Context, id string) (Customer, error)
	Create(ctx context.Context, c *Customer) error
	Update(ctx context.Context, c *Customer) error
	Delete(ctx context.Context, id string) error
}

type CustomerService interface {
	Index(ctx context.Context) ([]dto.CustomerData, error)
}
