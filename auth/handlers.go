package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"
)

func (api *API) Signup(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirm := c.FormValue("confirm")

	var usrCount = 0
	if api.postgress.Find(&Account{Email: email}).Count(&usrCount); usrCount > 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User with same email already exists")
	}

	if password != confirm {
		return echo.NewHTTPError(http.StatusUnauthorized, "Password must mutch confirmation")
	}

	acc, err := CreateAccount(api.postgress, email, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.HTML(http.StatusOK, fmt.Sprintf("User %s signed in", acc.ID))
}

func (api *API) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	acc, err := GetValidAccount(api.postgress, email, password)
	if err != nil {
		return echo.ErrUnauthorized
	}

	fmt.Println("User is logged in")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["expiration_date"] = time.Now().Add(time.Hour * time.Duration(api.config.JWT.ExpHours)).Unix()

	session := session.Default(c)
	session.Set("account_id", acc.ID)
	session.Save()

	t_string, err := token.SignedString([]byte(api.config.JWT.Secret))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"token": t_string})
}

func StubHandler(c echo.Context) error {
	return c.String(200, c.Request().URI())
}

func (api *API) GetRoles(c echo.Context) error {
	roles := &[]Role{}
	api.postgress.Find(roles)
	return c.JSON(http.StatusOK, roles)
}
