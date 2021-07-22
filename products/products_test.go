package products

import (
	"context"
	"fmt"
	"learning-golang-restful-api/config"
	"log"
	"net/http"
	"net/http/httptest"
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

func TestProduct(t *testing.T) {
	t.Run("Test create product", func(t *testing.T) {
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
	})

	t.Run("Test get products", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		ctx := e.NewContext(req, res)
		h.Coll = coll
		assert.Nil(t, h.GetProducts(ctx))
		assert.Equal(t, http.StatusOK, res.Code)
	})
}
