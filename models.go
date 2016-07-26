package testauth

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	gorm.Model
	Email    string
	password string
}

func (acc Account) Password() string {
	return acc.password
}

func (acc *Account) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("ID", uuid.New)
	return nil
}

func CreateAccount(db *gorm.DB, email, password string) (*Account, error) {
	account := &Account{}
	if !db.Where(&Account{Email: email}).First(account).RecordNotFound() {
		return nil, fmt.Errorf("Account with email: %s exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	account.Email = email
	account.password = string(hash)
	if err := db.Create(account).Error; err != nil {
		return nil, fmt.Errorf("Unable to create new account \r\n%v", err)
	}

	return account, nil
}

func IsValidAccount(db *gorm.DB, email, password string) (bool, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	account := &Account{}
	if db.Where(&Account{Email: email}).First(account).RecordNotFound() {
		return false, nil
	}

	return account.password == string(hash), nil
}

type User struct {
	gorm.Model
	Login       string
	Account     Account `gorm:"ForeignKey:AccountID"`
	AccountID   int
	Checkpoints []Checkpoint `gorm:"many2many:user_checkpoints;"`
	Predicates  []Predicate  `gorm:"many2many:role_predicates"`
	Roles       []Role       `gorm:"many2many:user_roles"`
}

type Role struct {
	gorm.Model
	Name         string
	ParentRole   *Role `gorm:"ForeignKey:ParentRoleID"`
	ParentRoleID int
	Checkpoints  []Checkpoint `gorm:"many2many:role_checkpoints"`
	Predicates   []Predicate  `gorm:"many2many:role_predicates"`
}

type Checkpoint struct {
	gorm.Model
	Name string
}

type Predicate struct {
	gorm.Model
	Name  string
	Query string
}

func CreateTables(db *gorm.DB) {
	db.CreateTable(&Account{}, &User{}, &Role{}, &Checkpoint{}, &Predicate{})
}
