package model

import (
	"gorm.io/gorm"
)

type Directory struct {
	gorm.Model
	Path     string `gorm:primaryKey`
	FileList []File //`gorm:"-"`
}
