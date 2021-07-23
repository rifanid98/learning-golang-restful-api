package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	C    *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
	h    *UsersHandler
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
	coll = db.Collection(cfg.DBUsersCollection)
	h = &UsersHandler{}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	testCode := m.Run()
	coll.Drop(ctx)
	db.Drop(ctx)
	os.Exit(testCode)
}

func TestUsers(t *testing.T) {
	t.Run("Test register invalid data", func(t *testing.T) {
		body := `
		{
			"username": "adninsijawa.office@gmail",
			"password": "password"
		}
		`
		req := httptest.NewRequest(http.MethodGet, "/auth", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		ctx := e.NewContext(req, res)
		h.Coll = coll
		t.Logf("res: %#+v\n", res.Body.Bytes())
		assert.Nil(t, h.RegisterUser(ctx))
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("Test register", func(t *testing.T) {
		t.Run("Test register with no is_admin", func(t *testing.T) {
			var user User
			body := `
		{
			"username": "adninsijawa.office@gmail.com",
			"password": "password"
		}
		`
			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
			res := httptest.NewRecorder()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			e := echo.New()
			ctx := e.NewContext(req, res)
			h.Coll = coll
			t.Logf("res: %#+v\n", res.Body.Bytes())
			assert.Nil(t, h.RegisterUser(ctx))
			assert.Equal(t, http.StatusCreated, res.Code)

			token := res.Header().Get("X-Auth-Token")
			assert.NotEmpty(t, token)

			err := json.Unmarshal(res.Body.Bytes(), &user)
			assert.Nil(t, err)
			assert.Equal(t, "adninsijawa.office@gmail.com", user.Email)
			assert.Empty(t, user.Password)
		})

		// t.Run("Test register with is_admin", func(t *testing.T) {
		// 	var user User
		// 	body := `
		// {
		// 	"username": "adninsijawa.medsos@gmail.com",
		// 	"password": "password",
		// 	"is_admin": true
		// }
		// `
		// 	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
		// 	res := httptest.NewRecorder()
		// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		// 	e := echo.New()
		// 	ctx := e.NewContext(req, res)
		// 	h.Coll = coll
		// 	t.Logf("res: %#+v\n", res.Body.Bytes())
		// 	assert.Nil(t, h.RegisterUser(ctx))
		// 	assert.Equal(t, http.StatusCreated, res.Code)

		// 	token := res.Header().Get("X-Auth-Token")
		// 	assert.NotEmpty(t, token)

		// 	err := json.Unmarshal(res.Body.Bytes(), &user)
		// 	assert.Nil(t, err)
		// 	assert.Equal(t, "adninsijawa.medsos@gmail.com", user.Email)
		// 	assert.Empty(t, user.Password)
		// })
	})

	// t.Run("Test register (again)", func(t *testing.T) {
	// 	body := `
	// 	{
	// 		"username": "adninsijawa.office@gmail.com",
	// 		"password": "password"
	// 	}
	// 	`
	// 	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	// 	res := httptest.NewRecorder()
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	e := echo.New()
	// 	ctx := e.NewContext(req, res)
	// 	h.Coll = coll
	// 	t.Logf("res: %#+v\n", res.Body.Bytes())
	// 	assert.NotNil(t, h.RegisterUser(ctx))
	// 	assert.Equal(t, http.StatusConflict, res.Code)
	// })

	// t.Run("Test login", func(t *testing.T) {
	// 	t.Run("Test login with authorized false", func(t *testing.T) {
	// 		// var user User
	// 		body := `
	// 	{
	// 		"username": "adninsijawa.office@gmail.com",
	// 		"password": "password"
	// 	}
	// 	`
	// 		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	// 		res := httptest.NewRecorder()
	// 		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 		e := echo.New()
	// 		ctx := e.NewContext(req, res)
	// 		h.Coll = coll
	// 		t.Logf("res: %#+v\n", res.Body.Bytes())
	// 		assert.Nil(t, h.LoginUser(ctx))
	// 		assert.Equal(t, http.StatusOK, res.Code)

	// 		token := res.Header().Get("X-Auth-Token")
	// 		assert.NotEmpty(t, token)
	// 		authorized := res.Header().Get("Au")
	// 		fmt.Println(authorized)
	// 		// assert.False(t, authorized.(bool))

	// 		// err := json.Unmarshal(res.Body.Bytes(), &user)
	// 		// assert.NotEmpty(t, err)
	// 		// assert.Equal(t, "adninsijawa.office@gmail.com", user.Email)
	// 		// assert.Empty(t, user.Password)
	// 	})
	// })
}
