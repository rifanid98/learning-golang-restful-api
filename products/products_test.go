package products

import (
	"context"
	"encoding/json"
	"fmt"
	"learning-golang-restful-api/config"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	C    *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
	cfg  config.Properties
	h    *ProductsHandler
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read : %v", err)
	}

	mongoURL := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)

	C, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("Unable to connect to database : %v", err)
	}

	db = C.Database(cfg.DBName)
	coll = db.Collection(cfg.DBCollection)
	h = &ProductsHandler{}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	testCode := m.Run()
	coll.Drop(ctx)
	db.Drop(ctx)
	os.Exit(testCode)
}

func TestProduct(t *testing.T) {
	var _ID string

	t.Run("Test create product", func(t *testing.T) {
		var ids []string
		body := `
		[
			{
				"name": "alexa",
				"price": 250,
				"currency": "USD",
				"vendor": "Amazon",
				"accessories": ["charger", "subscription"]
			}
		]
		`
		req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		ctx := e.NewContext(req, res)
		h.Coll = coll
		assert.Nil(t, h.CreateProducts(ctx))
		assert.Equal(t, http.StatusCreated, res.Code)

		err := json.Unmarshal(res.Body.Bytes(), &ids)
		assert.Nil(t, err)
		_ID = ids[0]
		t.Logf("IDs: %#+v\n", ids)
		for _, id := range ids {
			assert.NotNil(t, id)
		}

	})

	t.Run("Test get products", func(t *testing.T) {
		var products []Product
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		ctx := e.NewContext(req, res)
		h.Coll = coll
		assert.Nil(t, h.GetProducts(ctx))
		assert.Equal(t, http.StatusOK, res.Code)

		err := json.Unmarshal(res.Body.Bytes(), &products)
		assert.Nil(t, err)
		for _, product := range products {
			assert.Equal(t, "alexa", product.Name)
		}
	})

	t.Run("Test get products with query param", func(t *testing.T) {
		var products []Product
		req := httptest.NewRequest(http.MethodGet, "/products?currency=USD&vendor=Amazon", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		ctx := e.NewContext(req, res)
		h.Coll = coll
		assert.Nil(t, h.GetProducts(ctx))
		assert.Equal(t, http.StatusOK, res.Code)

		err := json.Unmarshal(res.Body.Bytes(), &products)
		assert.Nil(t, err)
		for _, product := range products {
			assert.Equal(t, "alexa", product.Name)
		}
	})

	t.Run("Test get a product", func(t *testing.T) {
		var product Product
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%s", _ID), nil)
		res := httptest.NewRecorder()
		e := echo.New()
		ctx := e.NewContext(req, res)
		ctx.SetParamNames("id")
		ctx.SetParamValues(_ID)
		h.Coll = coll
		err := h.GetProduct(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		err = json.Unmarshal(res.Body.Bytes(), &product)
		assert.Nil(t, err)
		assert.Equal(t, "USD", product.Currency)
	})

	t.Run("Test put a product", func(t *testing.T) {
		var product Product
		body := `
		{
			"name": "alexas",
			"price": 250,
			"currency": "USD",
			"vendor": "Amazon",
			"accessories": ["charger", "subscription"]
		}
		`
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/products/%s", _ID), strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		ctx := e.NewContext(req, res)
		ctx.SetParamNames("id")
		ctx.SetParamValues(_ID)
		h.Coll = coll
		err := h.UpdateProduct(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		err = json.Unmarshal(res.Body.Bytes(), &product)
		assert.Nil(t, err)
		assert.Equal(t, "USD", product.Currency)
	})

	t.Run("Test delete a product", func(t *testing.T) {
		var del int
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/products/%s", _ID), nil)
		res := httptest.NewRecorder()
		e := echo.New()
		ctx := e.NewContext(req, res)
		ctx.SetParamNames("id")
		ctx.SetParamValues(_ID)
		h.Coll = coll
		err := h.DeleteProduct(ctx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		err = json.Unmarshal(res.Body.Bytes(), &del)
		assert.Nil(t, err)
		assert.Equal(t, 1, del)
	})

}
