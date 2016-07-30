package auth

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// "github.com/pborman/uuid"
	"golang.org/x/crypto/bcrypt"
)

// type Cpt struct {
// CreateCpt   Checkpoint `gorm:"ForeignKey:CreateCptID"`
// CreateCptID uint
// ReadCpt     Checkpoint `gorm:"ForeignKey:ReadCptID"`
// ReadCptID   uint
// UpdateCpt   Checkpoint `gorm:"ForeignKey:UpdateCptID"`
// UpdateCptID uint
// DeleteCpt   Checkpoint `gorm:"ForeignKey:DeleteCptID"`
// DeleteCptID uint
// }

//Incapsulate gorm.Model and add CRUD checkpoints
// type CptModel struct {
// 	gorm.Model
// 	Cpt
// }

type Dictionary struct {
	gorm.Model
	Name        string
	Caption     string
	TableName   string
	CreateCpt   Checkpoint `gorm:"ForeignKey:CreateCptID"`
	CreateCptID uint
	ReadCpt     Checkpoint `gorm:"ForeignKey:ReadCptID"`
	ReadCptID   uint
	UpdateCpt   Checkpoint `gorm:"ForeignKey:UpdateCptID"`
	UpdateCptID uint
	DeleteCpt   Checkpoint `gorm:"ForeignKey:DeleteCptID"`
	DeleteCptID uint
}

type Account struct {
	gorm.Model
	Email    string
	Password string
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
	account.Password = string(hash)
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

	return account.Password == string(hash), nil
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
	Name      string
	ModelPath string //model_name from router path
	Query     string
}

type UserPredicateAction struct {
	gorm.Model
	UserID      uint `gorm:"primary_key:true"`
	PredicateID uint `gorm:"primary_key:true"`
	Create      bool `sql:"DEFAULT:false"`
	Read        bool `sql:"DEFAULT:false"`
	Update      bool `sql:"DEFAULT:false"`
	Delete      bool `sql:"DEFAULT:false"`
}

type RolePredicateAction struct {
	gorm.Model
	RoleID      uint `gorm:"primary_key:true"`
	PredicateID uint `gorm:"primary_key:true"`
	Create      bool `sql:"DEFAULT:false"`
	Read        bool `sql:"DEFAULT:false"`
	Update      bool `sql:"DEFAULT:false"`
	Delete      bool `sql:"DEFAULT:false"`
}

func DropTable(db *gorm.DB, model interface{}) {
	db.DropTableIfExists(model)
}

func CreateTable(db *gorm.DB, model interface{}) {
	s := db.CreateTable(model)
	tblName := s.NewScope(model).TableName()

	c_cpt := Checkpoint{Name: fmt.Sprintf("Checkpoint for Create dictionary \"%s\" record", tblName)}
	db.Create(&c_cpt)
	r_cpt := Checkpoint{Name: fmt.Sprintf("Checkpoint for Read dictionary \"%s\" record", tblName)}
	db.Create(&r_cpt)
	u_cpt := Checkpoint{Name: fmt.Sprintf("Checkpoint for Update dictionary \"%s\" record", tblName)}
	db.Create(&u_cpt)
	d_cpt := Checkpoint{Name: fmt.Sprintf("Checkpoint for Delete dictionary \"%s\" record", tblName)}
	db.Create(&d_cpt)

	d := &Dictionary{
		Name:        tblName,
		Caption:     tblName,
		TableName:   tblName,
		CreateCptID: c_cpt.ID,
		ReadCptID:   r_cpt.ID,
		UpdateCptID: u_cpt.ID,
		DeleteCptID: d_cpt.ID,
	}

	db.Create(d)
}

// func CreateTables(db *gorm.DB, models []*interface{}) {
// for _, model := range models {
// db.CreateTable(model)
// }
// }
