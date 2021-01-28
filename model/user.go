package model

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type status int

const (
	UNKNOWN status = iota // 0
	NORMAL
	DELETE
)

type User struct {
	ID     string `gorm:"primaryKey"`
	Pwd    string
	Status status `gorm:"default:0"`
}

func (u User) GetStatusName() string {
	switch u.Status {
	case UNKNOWN:
		return "UNKNOWN"
	case NORMAL:
		return "NORMAL"
	case DELETE:
		return "DELETE"
	default:
		return ""
	}

	return ""
}

func (u User) String() string {
	return fmt.Sprintf("[ID] : %s, [pwd] : %s, [status] : %s ", u.ID, u.Pwd, u.GetStatusName())
}

func (u User) HashPassword() (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Pwd), 14)
	return string(bytes), err
}

func (u User) CheckPasswordHash(pwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Pwd), []byte(pwd))
	return err == nil
}
