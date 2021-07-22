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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	c    *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
	cfg  config.Properties
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read : %v", err)
	}

	mongoURL := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)

	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("Unable to connect to database : %v", err)
	}

	db = c.Database(cfg.DBName)
	coll = db.Collection(cfg.DBCollection)
}

func main() {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	h := &products.ProductsHandler{Coll: coll}
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"))

	e.Logger.Infof("Listening on %s:%s ", cfg.AppHost, cfg.AppPort)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)))
}
