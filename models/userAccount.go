package models

import (
	"gopkg.in/mgo.v2/bson"
	"errors"
	"github.com/bitphinix/barbra_backend/db"
	"github.com/bitphinix/barbra_backend/helpers"
	"github.com/bitphinix/barbra_backend/payloads"
	"log"
)

var (
	ErrInvalidPayload    = errors.New("user: invalid payload")
	ErrEmailAlreadyInUse = errors.New("user: email already in use")
)

type UserAccount struct {
	Id        string               `json:"id"        bson:"_id"      validate:"hexadecimal" binding:"required"`
	Enrolled  bool                 `json:"enrolled"  bson:"enrolled"                        binding:"required"`
	Profile   *UserProfile         `json:"-"         bson:"profile"                         binding:"required"`
	Bookmarks map[string]*Bookmark `json:"-"         bson:"bookmarks"                       binding:"required"`
}

func GetUserAccount(id string) (*UserAccount, error) {
	collection := db.GetDB().C("users")

	account := new(UserAccount)
	err := collection.FindId(id).One(account)

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (account *UserAccount) IsEnrolled() bool {
	validate := helpers.GetValidator()
	err := validate.Struct(account)
	return err == nil
}

func RegisterUser(payload *payloads.ProfilePayload) (*UserAccount, error) {
	collection := db.GetDB().C("users")
	validate := helpers.GetValidator()

	//Check if payload is valid
	if err := validate.Struct(payload); err != nil {
		log.Println(err.Error())
		return nil, ErrInvalidPayload
	}

	//Check if email is already in use
	if payload.Email != "" {
		count, err := collection.Find(bson.M{"email": payload.Email}).Count()
		if err != nil || count > 0 {
			return nil, ErrEmailAlreadyInUse
		}
	}

	account := &UserAccount{
		Id:       bson.NewObjectId().Hex(),
		Enrolled: false,
		Profile: &UserProfile{
			Email:      payload.Email,
			FamilyName: payload.FamilyName,
			GivenName:  payload.GivenName,
			Nickname:   payload.Nickname,
			PictureURL: payload.PictureURL,
		},
		Bookmarks:make(map[string]*Bookmark),
	}

	account.Enrolled = account.IsEnrolled()
	return account, account.Save()
}

func (account *UserAccount) UpdateProfile(payload *payloads.ProfilePayload) error {
	err := account.Profile.UpdateInfo(payload)

	if err != nil {
		return err
	}

	account.Enrolled = account.IsEnrolled()
	return account.Update()
}


func (account *UserAccount) Save() error {
	collection := db.GetDB().C("users")
	return collection.Insert(account)
}

func (account *UserAccount) Update() error {
	collection := db.GetDB().C("users")
	return collection.UpdateId(account.Id, account)
}

func (account *UserAccount) Delete() error {
	collection := db.GetDB().C("users")
	return collection.RemoveId(account.Id)
}