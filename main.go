package main

import (
	"context"
	"fmt"
	"learning-golang-restful-api/config"
	"learning-golang-restful-api/products"
	"learning-golang-restful-api/users"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	C              *mongo.Client
	db             *mongo.Database
	cfg            config.Properties
	productsColl   *mongo.Collection
	usersColl      *mongo.Collection
	XCorrelationId = "X-Correlation-Id"
	XAuthToken     = "X-Auth-Token"
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
	productsColl = db.Collection(cfg.DBProductCollection)
	usersColl = db.Collection(cfg.DBUsersCollection)

	isUserIndexUnique := true
	indexModel := mongo.IndexModel{
		Keys: bson.D{{"username", 1}},
		Options: &options.IndexOptions{
			Unique: &isUserIndexUnique,
		},
	}

	_, err = usersColl.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Unable to create an index: %v", err)
	}
}

func addCorrelationId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// generate correlation id
		var newId string

		id := c.Request().Header.Get(XCorrelationId)
		if id == "" {
			// generate random number
			newId = random.String(12)
		} else {
			newId = id
		}

		c.Request().Header.Set(XCorrelationId, newId)
		c.Response().Header().Set(XCorrelationId, newId)
		return next(c)
	}
}

func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		hToken := c.Request().Header.Get("x-auth-token") // Bearer
		jwtToken := strings.Split(hToken, " ")[1]
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		})
		if err != nil {
			log.Errorf("Unable to parse token: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Unable to parse token")
		}
		if !claims["authorized"].(bool) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		return next(c)
	}
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.ERROR)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationId)
	e.Pre(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${header:X-Correlation-Id} ${host} ${method} ${uri} ${user_agent}` +
			`${status} ${error} ${latency_human}` + "\n",
	}))
	jwtConfig := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(cfg.JwtSecret),
		TokenLookup: "header:" + XAuthToken,
		AuthScheme:  "Bearer",
	})

	h := &products.ProductsHandler{Coll: productsColl}
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"), jwtConfig)
	e.GET("/products", h.GetProducts, jwtConfig)
	e.GET("/products/:id", h.GetProduct, jwtConfig)
	e.PUT("/products/:id", h.UpdateProduct, middleware.BodyLimit("1M"), jwtConfig)
	e.DELETE("/products/:id", h.DeleteProduct, jwtConfig, adminMiddleware)

	uh := &users.UsersHandler{Coll: usersColl}
	e.POST("/auth/register", uh.RegisterUser)
	e.POST("/auth/login", uh.LoginUser)
	e.Logger.Infof("Listening on %s:%s ", cfg.AppHost, cfg.AppPort)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)))
}
