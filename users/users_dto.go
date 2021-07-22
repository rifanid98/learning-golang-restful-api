package users

import (
	"github.com/go-playground/validator"
)

var (
	v = validator.New()
)

// User describes an electronic product e.g. phone
type User struct {
	Email    string `json:"username" bson:"username" validate:"required,email"`
	Password string `json:"password" bson:"password" validate:"required,min=8,max=300"`
}

// UserValidator a product validator
type UserValidator struct {
	validator *validator.Validate
}

// Validates a product
func (v *UserValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
