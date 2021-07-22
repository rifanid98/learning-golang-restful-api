package products

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

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

func findProducts(ctx context.Context, qs url.Values, collection CollectionAPI) ([]Product, error) {
	var products []Product
	filter := make(map[string]interface{})
	for k, v := range qs {
		filter[k] = v[0]
	}

	if filter["_id"] != nil {
		// convert string value of _id to ObjectID
		_id, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			return products, err
		}
		filter["_id"] = _id
	}

	cursor, err := collection.Find(ctx, bson.M(filter))
	if err != nil {
		log.Errorf("Unable to find products : %v", err)
		return products, nil
	}

	err = cursor.All(ctx, &products)
	if err != nil {
		log.Errorf("Unable to read the cursor : %v", err)
		return products, nil
	}

	return products, nil
}

func (h *ProductsHandler) GetProducts(c echo.Context) error {
	products, err := findProducts(context.Background(), c.QueryParams(), h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &products)
}

func findProduct(ctx context.Context, id string, coll CollectionAPI) (*Product, error) {
	var product Product

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to convert id to _id: %v", err)
		return &product, err
	}

	filter := bson.M{"_id": _id}
	res := coll.FindOne(ctx, filter)
	if err := res.Decode(&product); err != nil {
		log.Errorf("Unable to decode FindOne res: %v", err)
		return &product, err
	}

	return &product, nil
}

func (h *ProductsHandler) GetProduct(c echo.Context) error {
	products, err := findProduct(context.Background(), c.Param("id"), h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &products)
}

func updateProduct(ctx context.Context, id string, body io.ReadCloser, coll CollectionAPI) (*Product, error) {
	var product Product

	// 1. find product or return 404
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to convert id to _id: %v", err)
		return &product, err
	}

	filter := bson.M{"_id": _id}
	res := coll.FindOne(ctx, filter)
	if err := res.Decode(&product); err != nil {
		log.Errorf("Unable to decode FindOne res: %v", err)
		return &product, err
	}

	// 2. decode body to struct or return 500
	if err := json.NewDecoder(body).Decode(&product); err != nil {
		log.Errorf("Unable to decode from req body to struct: %v", err)
		return &product, err
	}

	// 3. validate decoded body or return 400
	if err := v.Struct(&product); err != nil {
		log.Errorf("Unable to decode from struct: %v", err)
		return &product, err
	}

	// 4. update data or return 500
	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": product})
	if err != nil {
		log.Errorf("Unable to update: %v", err)
		return &product, err
	}

	return &product, nil
}

func (h *ProductsHandler) UpdateProduct(c echo.Context) error {
	c.Echo().Validator = &ProductValidator{validator: v}

	product, err := updateProduct(context.Background(), c.Param("id"), c.Request().Body, h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, product)
}

func deleteProduct(ctx context.Context, id string, coll CollectionAPI) (int, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to convert id to _id: %v", err)
	}

	res, err := coll.DeleteOne(ctx, bson.M{"_id": _id})
	if err != nil {
		log.Errorf("Unable to delete data: %v", err)
	}

	return int(res.DeletedCount), nil
}

func (h *ProductsHandler) DeleteProduct(c echo.Context) error {
	del, err := deleteProduct(context.Background(), c.Param("id"), h.Coll)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, del)
}
