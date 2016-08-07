package auth

import (
	"fmt"
	"sort"

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
func NewAPI(config Config) (*API, error) {
	api := &API{
		config: &config,
		router: echo.New(),
	}

	var err error
	api.postgress, err = gorm.Open("postgres", config.PostgeressConString)
	if err != nil {
		panic(err)
	}

	api.router.Use(middleware.Logger())
	api.router.Use(middleware.Recover())

	api.setRoutes()

	jwtConfig := middleware.DefaultJWTConfig
	jwtConfig.SigningKey = []byte(config.JWT.Secret)
	jwtConfig.Skipper = func(c echo.Context) bool {
		uri := c.Request().URI()
		result := sort.SearchStrings(config.JWT.ExcludedPaths, uri) >= 0
		return result
	}

	api.router.Use(middleware.JWTWithConfig(jwtConfig))

	store := session.NewCookieStore([]byte("secret"))
	api.router.Use(session.Sessions("GSESSION", store))

	if api.postgress == nil {
		fmt.Println("POSTGRESS IS NILL")
	}
	api.router.Use(CheckRightAccess(api.postgress))
	return api, nil
}

func (api *API) createTbles() {
	api.postgress.DropTableIfExists(&Account{}, &Role{}, &Predicate{}, &AccountPredicateAction{}, &RolePredicateAction{})
	//CreateTable(api.postgress, &Dictionary{})
	CreateDictionary(api.postgress, &Account{})
	CreateDictionary(api.postgress, &Role{})
	CreateDictionary(api.postgress, &Predicate{})
	CreateDictionary(api.postgress, &AccountPredicateAction{})
	CreateDictionary(api.postgress, &RolePredicateAction{})
}

func (api *API) Run(srv engine.Server) {
	//api.CreateTbles()
	// checkpoints := []Checkpoint{}
	// api.postgress.Find(&checkpoints)

	// admin := &Role{Name: "admin"}
	// admin.Checkpoints = checkpoints

	// api.postgress.Create(admin)
	api.router.Run(srv)
}
