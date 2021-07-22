package products

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductsHandler struct {
	Coll CollectionAPI
}

func createProducts(ctx context.Context, products []Product, coll CollectionAPI) ([]interface{}, error) {
	var insertedIds []interface{}

	for _, product := range products {
		product.ID = primitive.NewObjectID()

		res, err := coll.InsertOne(ctx, product)
		if err != nil {
			log.Fatalf("Unable to insert: %v", err)
			return nil, err
		}

		insertedIds = append(insertedIds, res.InsertedID)
	}

	return insertedIds, nil
}

func (h *ProductsHandler) CreateProducts(c echo.Context) error {
	var products []Product
	if err := c.Bind(&products); err != nil {
		log.Fatalf("Unable to bind: %v", err)
	}

	ids, err := createProducts(context.Background(), products, h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, ids)
}
