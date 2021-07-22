package products

import (
	"context"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
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
			log.Errorf("Unable to insert: %v", err)
			return nil, err
		}

		insertedIds = append(insertedIds, res.InsertedID)
	}

	return insertedIds, nil
}

func (h *ProductsHandler) CreateProducts(c echo.Context) error {
	c.Echo().Validator = &ProductValidator{validator: v}

	var products []Product
	if err := c.Bind(&products); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return err
	}

	for _, product := range products {
		if err := c.Validate(product); err != nil {
			log.Errorf("Unable to validate the product %+v %v", product, err)
			return err
		}
	}

	ids, err := createProducts(context.Background(), products, h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, ids)
}

func findProducts(ctx context.Context, collection CollectionAPI) ([]Product, error) {
	var products []Product
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Errorf("Unable to find products : %v", err)
	}

	err = cursor.All(ctx, &products)
	if err != nil {
		log.Errorf("Unable to read the cursor : %v", err)
	}

	return products, nil
}

func (h *ProductsHandler) GetProducts(c echo.Context) error {
	products, err := findProducts(context.Background(), h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &products)
}
