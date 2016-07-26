package main

import (
	// "log"
	"net/http"
	// "time"
	"fmt"
	"io/ioutil"

	// "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"github.com/rpoletaev/testauth"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                string `yaml:"port"`
	PostgeressConString string `yaml:"connection_string"`
}

type API struct {
	router    *echo.Echo
	postgress *gorm.DB

	config *Config
}

func (api *API) setRoutes() {
	api.router.POST("login", func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")
		if ok, err := IsValidAccount(api.postgress, email, password); ok {
			return c.String(http.StatusOK, "User is valid!!!")
		} else {
			return c.String(http.StatusMethodNotAllowed, err.Error())
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

	var err error
	api.postgress, err = gorm.Open("postgress", config.PostgeressConString)
	if err != nil {
		panic(err)
	}

	return api, nil
}

func main() {
	config := &Config{}
	configBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err)
		config.Port = ":3030"
		config.PostgeressConString = ""
	}

	err = yaml.Unmarshal(configBytes, config)
	if err != nil {

	}

	api, erro := NewApi(*config)
	if erro != nil {
		fmt.Println(erro)
	}

	api.router.Run(standard.New(config.Port))
}
