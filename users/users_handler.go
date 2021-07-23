package users

import (
	"context"
	"fmt"
	"learning-golang-restful-api/config"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	cfg config.Properties
)

type UsersHandler struct {
	Coll CollectionAPI
}

func createUser(ctx context.Context, user User, coll CollectionAPI) (interface{}, *echo.HTTPError) {
	var newUser User
	findRes := coll.FindOne(ctx, bson.M{"username": user.Email})
	err := findRes.Decode(&newUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("Unable to decode retrieved user: %s", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to decode retrieved user")
	}

	if newUser.Email != "" {
		log.Errorf("User by %s already exists", newUser.Email)
		return nil, echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		log.Errorf("Unable to hash the password: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to process the password")
	}

	user.Password = string(pass)

	insertRes, err := coll.InsertOne(ctx, user)
	if err != nil {
		log.Errorf("Unable to insert: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to insert")
	}

	return insertRes.InsertedID, nil
}

func (h *UsersHandler) RegisterUser(c echo.Context) error {
	c.Echo().Validator = &UserValidator{validator: v}

	var user User
	if err := c.Bind(&user); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to bind data")
	}

	if err := c.Validate(user); err != nil {
		log.Errorf("Unable to validate the user %+v %v", user, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to validate the user")
	}

	ids, err := createUser(context.Background(), user, h.Coll)
	if err != nil {
		return err
	}

	token, tokenErr := user.createToken()
	if tokenErr != nil {
		log.Errorf("Unable to generate the token: %v", tokenErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}

	c.Response().Header().Set("x-auth-token", "Bearer "+token)
	return c.JSON(http.StatusCreated, ids)
}

func isCredValid(givenPassword, hashedPassword string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(givenPassword)); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (u User) createToken() (string, error) {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read: %v", err)
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = u.IsAdmin
	claims["user_id"] = u.Email
	claims["exp"] = time.Now().Add(15 * time.Minute).Unix()

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(cfg.JwtSecret))
	if err != nil {
		log.Errorf("Unable to generate the token: %v", err)
	}

	return token, nil
}

func loginUser(ctx context.Context, user *User, coll CollectionAPI) (interface{}, *echo.HTTPError) {
	givenPassword := user.Password

	res := coll.FindOne(ctx, bson.M{"username": user.Email})
	err := res.Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("Unable to decode retrieved user: %s", err)
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, "Unable to decode retrieved user")
	}

	if err == mongo.ErrNoDocuments {
		log.Errorf("Unable %s does not exists", user.Email)
		return nil, echo.NewHTTPError(http.StatusNotFound, "User does not exists")
	}

	if !isCredValid(givenPassword, user.Password) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid Credentials")
	}

	return User{Email: user.Email}, nil
}

func (h *UsersHandler) LoginUser(c echo.Context) error {
	c.Echo().Validator = &UserValidator{validator: v}

	var user User
	if err := c.Bind(&user); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Unable to bind data")
	}

	if err := c.Validate(user); err != nil {
		log.Errorf("Unable to validate the user %+v %v", user, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to validate the payload")
	}

	ids, err := loginUser(context.Background(), &user, h.Coll)
	if err != nil {
		return err
	}

	token, tokenErr := user.createToken()
	if tokenErr != nil {
		log.Errorf("Unable to generate the token: %v", tokenErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate the token")
	}

	c.Response().Header().Set("x-auth-token", "Bearer "+token)
	return c.JSON(http.StatusCreated, ids)
}
