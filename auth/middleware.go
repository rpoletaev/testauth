package auth

import (
	"fmt"
	//"github.com/dgrijalva/jwt-go"
	"github.com/ipfans/echo-session"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	//"strconv"
	"strings"
)

func CheckRightAccess(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			exludedResources := []string{"/", "/login", "/signup", "/favicon.ico"}
			for _, resource := range exludedResources {
				// fmt.Println("uri is ", c.Request().URI(), " resource is ", resource)
				if resource == c.Request().URI() {
					// fmt.Println("email is", c.FormValue("email"))
					return next(c)
				}
			}

			var accountID uint
			session := session.Default(c)

			//try get account from session
			accountID = session.Get("account_id").(uint)

			acc := &Account{}
			db.First(acc, accountID)
			if acc == nil {
				return fmt.Errorf("Account %d isn't exist!", accountID)
			}

			//try get dictionary from URI
			dict := GetDictionaryFromUri(db, c.Request().URI())
			if &dict == nil {
				return fmt.Errorf("Dictionary [%s] isn't exist", dict.Name)
			}

			method := c.Request().Method()
			action := getActionSynonim(method)
			if !acc.HasActionAccess(db, dict, method) {
				return fmt.Errorf("User [%d] hasn't checkpoint to %s [%s]", accountID, action, dict.Name)
			}

			fmt.Printf("User [%d] has access to %s [%s]\n", accountID, action, dict.Name)

			predicates := acc.GetPredicatesForDictActions(db, dict, method)
			allowedList := []uint{}
			for _, predicate := range predicates {
				allowedList = append(allowedList, predicate.AllowedList(db, accountID)...)
			}

			fmt.Println("Allowed list containts ", allowedList)

			return next(c)
		}
	}
}

func GetDictionaryFromUri(db *gorm.DB, uri string) Dictionary {
	splitURI := strings.Split(uri, "/")
	name := splitURI[1]
	println("Splited uri ", uri, " ", name)
	dictionary := Dictionary{}
	db.Where("name = ?", name).First(&dictionary)
	return dictionary
}

func getActionSynonim(action string) string {
	switch strings.ToUpper(action) {
	case "GET":
		return "read"
	case "PUT":
		return "update"
	case "POST":
		return "create"
	case "DELETE":
		return "delete"
	}

	return "uncnown action"
}
