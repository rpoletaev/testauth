package auth

import (
	"github.com/ipfans/echo-session"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/middleware"
)

type Config struct {
	Port                string `yaml:"port"`
	PostgeressConString string `yaml:"connection_string"`
	JWT                 struct {
		Secret        string   `yaml:"secret"`
		ExpHours      int      `yaml:"exp_hours"`
		ExcludedPaths []string `yaml:"excluded_path"` //URI которые не проверяют jwt токен
	}
}

type API struct {
	router    *echo.Echo
	postgress *gorm.DB
	config    *Config
}

func (api *API) setRoutes() {
	api.router.POST("/login", api.Login)
	api.router.GET("/login", StubHandler)
	api.router.POST("/signup", api.Signup)
	api.router.Any("/", StubHandler)
	api.router.GET("/roles", api.GetRoles)
}

//NewAPI Create & configure API object
func NewAPI(config Config, init bool) (*API, error) {
	api := &API{
		config: &config,
		router: echo.New(),
	}

	var err error
	api.postgress, err = gorm.Open("postgres", config.PostgeressConString)

	if err != nil {
		panic(err)
	}

	if init {
		api.initialization()
	}

	api.router.Use(middleware.Logger())
	api.router.Use(middleware.Recover())

	api.setRoutes()

	jwtConfig := middleware.DefaultJWTConfig
	jwtConfig.SigningKey = []byte(config.JWT.Secret)
	jwtConfig.Skipper = func(c echo.Context) bool {
		uri := c.Request().URI()
		for _, resource := range config.JWT.ExcludedPaths {
			if resource == uri {
				println("resource is ", resource, "JWT WAS ESCAPED")
				return true
			}
		}
		return false
	}

	api.router.Use(middleware.JWTWithConfig(jwtConfig))

	store := session.NewCookieStore([]byte("secret"))
	api.router.Use(session.Sessions("GSESSION", store))

	api.router.Use(CheckRightAccess(api.postgress))
	return api, nil
}

func (api *API) initialization() {
	api.postgress.DropTableIfExists(&Dictionary{}, &Checkpoint{}, &Account{}, &Role{}, &Predicate{}, &AccountPredicateAction{}, &RolePredicateAction{})
	api.postgress.CreateTable(&Dictionary{}, &Checkpoint{}, &Role{})
	api.postgress.Create(&Role{Name: "admin"})
	CreateDictionary(api.postgress, &Dictionary{}, "Справочники")
	CreateDictionary(api.postgress, &Checkpoint{}, "Контрольные точки")
	CreateDictionary(api.postgress, &Role{}, "Роли")
	CreateDictionary(api.postgress, &Account{}, "Аккаунты")
	CreateDictionary(api.postgress, &AccountPredicateAction{}, "Связь акаунтов, предикатов и действий")
	CreateDictionary(api.postgress, &RolePredicateAction{}, "Связь ролей, предикатов и действий")
}

func (api *API) Run(srv engine.Server) {
	api.router.Run(srv)
}
