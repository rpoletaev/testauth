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
	if email == "" || password == "" || confirm == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "email, password and confirmation is required")
	}

	if password != confirm {
		return echo.NewHTTPError(http.StatusUnauthorized, "Password must mutch confirmation")
	}

	acc, err := CreateAccount(api.postgress, email, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.HTML(http.StatusCreated, fmt.Sprintf("User %s signed in", acc.ID))
}

func (api *API) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	acc, err := GetValidAccount(api.postgress, email, password)
	if err != nil {
		return echo.ErrUnauthorized
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(api.config.JWT.ExpHours)).Unix()

	t, err := token.SignedString([]byte(api.config.JWT.Secret))
	if err != nil {
		return err
	}

	//Response().Header().Add(echo.HeaderAuthorization, "Bearer "+t)
	session := session.Default(c)
	session.Set("account_id", acc.ID)
	session.Save()

	return c.JSON(http.StatusOK, map[string]string{"token": t})
}

func StubHandler(c echo.Context) error {
	return c.String(200, c.Request().URI())
}

func (api *API) GetRoles(c echo.Context) error {
	roles := &[]Role{}
	api.postgress.Find(roles)
	return c.JSON(http.StatusOK, roles)
}
