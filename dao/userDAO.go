package dao

import (
	"fmt"
	"tutorial/model"

	"github.com/pkg/errors"

	"gorm.io/gorm"
)

// mockgen -destination=./mocks/mock_userDAO.go -package=mocks -source=./dao/userDAO.go
type UserDAO interface {
	Get(id string) (*model.User, error)
	Post(*model.User) error
	Delete(id string) error
	SetDatabase(db *gorm.DB)
}

type UserData struct {
	db *gorm.DB
}

func (u UserData) Get(id string) (*model.User, error) {
	var user model.User

	tx := u.db.First(&user, "id = ?", id)

	if tx.Error != nil {
		return nil, errors.Wrapf(tx.Error, "failed to query User: %q", id)
	}

	return &user, nil
}

func (u UserData) Delete(id string) error {
	fmt.Println("UserData.Delete() start...")
	delUser, err := u.Get(id)

	if err != nil {
		return err
	}

	result := u.db.Model(&delUser).Update("status", model.DELETE)

	fmt.Println("UserData.Delete() end...")

	return result.Error
}

func DeleteAfterTester(u UserDAO) {
	fmt.Println("UserData.DeleteAfterTester() start...")

	user := &model.User{ID: "admin", Pwd: "1234", Status: 1}

	if err := u.Post(user); err != nil {
		fmt.Printf("Post(%q) = %q \n", user.String(), err.Error())
	}

	fmt.Printf("Success Post(%q) \n", user.String())

	getUser, err := u.Get(user.ID)

	if err != nil {
		fmt.Printf("Get(%q) = %q \n", user.ID, err)
	}

	fmt.Printf("Success Get(%q) = %q \n", user.ID, getUser.String())

	if err := u.Delete(user.ID); err != nil {
		fmt.Printf("Delete(%q) = %q \n", user.ID, err)
	}

	fmt.Printf("Success Delete(%q) \n", user.ID)

	fmt.Println("UserData.DeleteAfterTester() end...")
}

func (u UserData) Post(user *model.User) error {
	result := u.db.Create(user)

	return result.Error
}

func (u *UserData) SetDatabase(db *gorm.DB) {
	u.db = db
}
