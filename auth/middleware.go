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
			db.AutoMigrate(&Predicate{})
			exludedResources := []string{"/", "/login", "/signup"}
			for _, resource := range exludedResources {
				if resource == c.Request().URI() {
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
			if !acc.HasActionAccess(db, dict, method) {
				return fmt.Errorf("User [%d] hasn't checkpoint to Dictionary [%s]", accountID, dict.Name)
			}

			fmt.Printf("User [%d] has access to dictionary [%s]\n", accountID, dict.Name)

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

	dictionary := Dictionary{}
	db.Where("name = ?", name).First(&dictionary)
	return dictionary
}
