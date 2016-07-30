package main

import (
	// "log"
	"net/http"
	// "time"
	"fmt"
	"io/ioutil"

	// "github.com/dgrijalva/jwt-go"
	"github.com/ipfans/echo-session"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"github.com/rpoletaev/testauth/auth"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                string `yaml:"port"`
	PostgeressConString string `yaml:"connection_string"`
}

type API struct {
	router    *echo.Echo
	postgress *gorm.DB
	config    *Config
}

func (api *API) setRoutes() {
	api.router.POST("/login", func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if ok, err := auth.IsValidAccount(api.postgress, email, password); ok {
			return c.JSON(http.StatusOK, "User is valid!!!")
		} else {
			if err != nil {
				fmt.Println(err.Error())
			}
			return c.Redirect(http.StatusForbidden, "/signup")
		}
	})

	api.router.POST("/signup", func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")
		confirm := c.FormValue("confirm")

		acc := &auth.Account{}
		if api.postgress.Count(&Account{Email: email}) > 0 {
			return echo.NewHTTPError(http.StatusUnauthorized, "User with same email already exists")
		}

		if password != confirm {
			return echo.NewHTTPError(http.StatusUnauthorized, "Password must mutch confirmation")
		}
	})
}

func NewApi(config Config) (*API, error) {
	api := &API{
		config: &config,
		router: echo.New(),
	}

	api.router.Use(middleware.Logger())
	api.router.Use(middleware.Recover())
	api.setRoutes()
	var err error
	api.postgress, err = gorm.Open("postgres", config.PostgeressConString)
	if err != nil {
		panic(err)
	}

	api.postgress.DropTableIfExists(&auth.Account{}, &auth.User{}, &auth.Role{}, &auth.Predicate{}, &auth.UserPredicateAction{}, &auth.RolePredicateAction{})
	auth.CreateTable(api.postgress, &auth.Account{})
	auth.CreateTable(api.postgress, &auth.User{})
	auth.CreateTable(api.postgress, &auth.Role{})
	auth.CreateTable(api.postgress, &auth.Checkpoint{})
	auth.CreateTable(api.postgress, &auth.UserPredicateAction{})
	auth.CreateTable(api.postgress, &auth.RolePredicateAction{})
	// auth.CreateTables(api.postgress)
	return api, nil
}

func main() {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		panic(err)
	}

	api, erro := NewApi(*config)
	if erro != nil {
		fmt.Println(erro)
	}

	api.router.Run(standard.New(config.Port))
}
