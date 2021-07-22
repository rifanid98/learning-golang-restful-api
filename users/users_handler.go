package users

import (
	"context"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	insertRes, err := coll.InsertOne(ctx, user)
	if err != nil {
		log.Errorf("Unable to insert: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to insert")
	}

	return insertRes.InsertedID, nil
}

func (h *UsersHandler) CreateUser(c echo.Context) error {
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

	return c.JSON(http.StatusCreated, ids)
}
