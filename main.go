package main

import (
	"context"
	"fmt"
	"learning-golang-restful-api/config"
	"learning-golang-restful-api/products"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	C             *mongo.Client
	db            *mongo.Database
	coll          *mongo.Collection
	cfg           config.Properties
	CorrelationId = "X-Correlation-Id"
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
}

func addCorrelationId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// generate correlation id
		var newId string

		id := c.Request().Header.Get(CorrelationId)
		if id == "" {
			// generate random number
			newId = random.String(12)
		} else {
			newId = id
		}

		c.Request().Header.Set(CorrelationId, newId)
		c.Response().Header().Set(CorrelationId, newId)
		return next(c)
	}
}

func main() {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationId)

	h := &products.ProductsHandler{Coll: coll}
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"))

	e.Logger.Infof("Listening on %s:%s ", cfg.AppHost, cfg.AppPort)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)))
}
