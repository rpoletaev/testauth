package auth

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// "github.com/pborman/uuid"
	"strings"

	"golang.org/x/crypto/bcrypt"
	// "container/list"
)

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
	Email       string
	Password    string
	Checkpoints []Checkpoint `gorm:"many2many:account_checkpoints"`
	Predicates  []Predicate  `gorm:"many2many:account_predicates"`
	Roles       []Role       `gorm:"many2many:account_roles"`
}

func CreateAccount(db *gorm.DB, email, password string) (*Account, error) {
	account := &Account{}
	withSameEmail := 0

	// db.Find(&[]Account{}, &Account{Email: email}).Count(&withSameEmail)
	db.Model(&Account{}).Where("email = ?", email).Count(&withSameEmail)
	println(withSameEmail)

	if withSameEmail > 0 {
		return nil, fmt.Errorf("Account with email: %s exist", email)
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

	cnt := 0
	db.Find(&[]Account{}).Count(&cnt)

	if cnt == 1 {
		admin := &Role{}
		db.First(admin, 1)
		db.Model(account).Association("Roles").Append(admin)
	}
	return account, nil
}

func GetValidAccount(db *gorm.DB, email, password string) (*Account, error) {
	account := &Account{}
	if db.Where(&Account{Email: email}).First(account).RecordNotFound() {
		return nil, fmt.Errorf("Record not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, err
	}

	return account, nil
}

func (acc *Account) HasActionAccess(db *gorm.DB, dict Dictionary, action string) (result bool) {
	switch strings.ToUpper(action) {
	default:
		result = false
	case "GET":
		result = acc.IsCheckpointAllowed(db, dict.ReadCptID)
	case "PUT":
		result = acc.IsCheckpointAllowed(db, dict.UpdateCptID)
	case "POST":
		result = acc.IsCheckpointAllowed(db, dict.CreateCptID)
	case "DELETE":
		result = acc.IsCheckpointAllowed(db, dict.DeleteCptID)
	}

	return
}

func (acc *Account) IsCheckpointAllowed(db *gorm.DB, cpt uint) bool {
	checkpoints := []Checkpoint{}
	db.Model(acc).Association("Checkpoints").Find(&checkpoints)

	for _, cp := range checkpoints {
		if cp.ID == cpt {
			return true
		}
	}

	accRoles := []Role{}
	db.Model(acc).Association("Roles").Find(&accRoles)

	for _, role := range accRoles {
		if role.IsCheckpointAllowed(db, cpt) {
			return true
		}
	}
	return false
}

func (acc *Account) GetPredicatesForDictActions(db *gorm.DB, dict Dictionary, action string) []Predicate {
	actionCondition := getActionCondition(action)
	if actionCondition != "" {
		actionCondition = " and " + actionCondition
	}

	accountPredicates := []Predicate{}
	predicates := []Predicate{}
	var predicateIds []uint
	db.Model(&AccountPredicateAction{}).Where("account_id = ? "+actionCondition, acc.ID).Pluck("predicate_id", &predicateIds)

	if len(predicateIds) > 0 {
		db.Model(acc).Related(&predicates, "Predicates").Find(&accountPredicates, "dictionary_id = ? and id in (?)", dict.ID, predicateIds)
	}

	rolePredicates := []Predicate{}
	roles := []Role{}
	roleIds := []uint{}
	db.Model(acc).Related(&roles, "Roles").Pluck("id", &roleIds)
	db.Model(&RolePredicateAction{}).Where("role_id in (?) "+actionCondition, roleIds).Pluck("predicate_id", &predicateIds)
	if len(predicateIds) > 0 {
		db.Model(&Role{}).Related(&predicates, "Predicates").Find(rolePredicates, "id in (?) and dictionary_id = ?", predicateIds, dict.ID)
	}

	fmt.Println("Account with role predicates has ", len(rolePredicates), " items")

	predicates = append(accountPredicates, rolePredicates...)
	return predicates
}

type Role struct {
	gorm.Model
	Name         string
	ParentRole   *Role `gorm:"ForeignKey:ParentRoleID"`
	ParentRoleID int
	Checkpoints  []Checkpoint `gorm:"many2many:role_checkpoints"`
	Predicates   []Predicate  `gorm:"many2many:role_predicates"`
}

func (r *Role) IsCheckpointAllowed(db *gorm.DB, cptId uint) bool {
	cpoints := []Checkpoint{}
	db.Model(r).Association("Checkpoints").Find(&cpoints)

	for _, cp := range cpoints {
		if cp.ID == cptId {
			return true
		}
	}
	return false
}

type Checkpoint struct {
	gorm.Model
	Name string
}

func (cpt *Checkpoint) AfterCreate(tx *gorm.DB) (err error) {
	admin := &Role{}
	tx.First(admin, 1)
	tx.Model(admin).Association("Checkpoints").Append(*cpt)
	tx.Save(admin)
	return
}

//Predicate contains display name like "Only User Roles"
//Query - Raw Sql query with only one parameter account_id. Query must select only one field "id"
type Predicate struct {
	gorm.Model
	Name         string
	Query        string
	Dictionary   Dictionary
	DictionaryID uint
}

//AllowedList returns only allowed id from predicate dictionary
func (predicate *Predicate) AllowedList(db *gorm.DB, accId uint) []uint {
	list := []uint{}
	db.Raw(predicate.Query, accId).Pluck("id", &list)
	return list
}

type AccountPredicateAction struct {
	gorm.Model
	AccountID   uint `gorm:"primary_key:true"`
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

//Create tables for model and assign CRUD checkpoints
func CreateDictionary(db *gorm.DB, model interface{}, caption string) {
	var tblName string
	if db.HasTable(model) {
		tblName = db.Model(model).NewScope(model).TableName()
	} else {
		s := db.CreateTable(model)
		tblName = s.NewScope(model).TableName()
	}

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
		Caption:     caption,
		TableName:   tblName,
		CreateCptID: c_cpt.ID,
		ReadCptID:   r_cpt.ID,
		UpdateCptID: u_cpt.ID,
		DeleteCptID: d_cpt.ID,
	}

	db.Create(d)
}

func getActionCondition(action string) string {
	switch strings.ToUpper(action) {
	case "GET":
		return "read = true"
	case "PUT":
		return "update = true"
	case "POST":
		return "create = true"
	case "DELETE":
		return "delete = true"
	}

	return ""
}
