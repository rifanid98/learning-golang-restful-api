package products

import (
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	v = validator.New()
)

// Product describes an electronic product e.g. phone
type Product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name" validate:"required,max=10"`
	Price       int                `json:"price" bson:"price" validate:"required,max=2000"`
	Currency    string             `json:"currency" bson:"currency" validate:"required,len=3"`
	Discount    int                `json:"discount" bson:"discount"`
	Vendor      string             `json:"vendor" bson:"vendor" validate:"required"`
	Accessories []string           `json:"accessories,omitempty" bson:"accessories,omitempty" validate:"required"`
	IsEssential bool               `json:"is_essential" bson:"is_essential"`
}

// ProductValidator a product validator
type ProductValidator struct {
	validator *validator.Validate
}

// Validates a product
func (v *ProductValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
