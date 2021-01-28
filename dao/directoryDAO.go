package dao

import (
	"tutorial/model"

	"github.com/pkg/errors"

	"gorm.io/gorm"
)

type DirectoryList []model.Directory

type DirectoryDAO interface {
	CreateDirectory(path string) error
	DeleteDirectory(dirPath string, filePath string) error
	ModifyDirectory(fromDir string, toDir string, fromFile string, toFile string) error
	LoadDirectoryList(path string) (*DirectoryList, error)
	LoadDirectory(path string) (*model.Directory, error)
	SetDatabase(db *gorm.DB)
	GetDatabase() *gorm.DB
}

type DirectoryData struct {
	db *gorm.DB
}

func (d DirectoryData) CreateDirectory(path string) error {
	dir := new(model.Directory)

	dir.Path = path

	return d.db.Create(dir).Error
}

func (d DirectoryData) DeleteDirectory(dirPath string, filePath string) error {
	if err := d.db.Where("path = ?", filePath).Delete(&model.File{}).Error; err != nil {
		return errors.Wrapf(err, "디렉터리 내부 파일 삭제 실패 %q", filePath)
	}

	if err := d.db.Where("path = ?", dirPath).Delete(&model.Directory{}).Error; err != nil {
		return errors.Wrapf(err, "디렉터리 삭제 실패 %q", dirPath)
	}

	return nil
}

func (d DirectoryData) modify(fromDir string, toDir string, fromFile string, toFile string) func(tx *gorm.DB) error {
	return func(tx *gorm.DB) error {
		if err := tx.Model(model.File{}).Where("path = ?", fromFile).Updates(model.File{Path: toFile}).Error; err != nil {
			return errors.Wrapf(err, "디렉터리 내부 파일 수정 실패 %q", fromFile)
		}

		if err := tx.Model(model.Directory{}).Where("path = ?", fromDir).Updates(model.Directory{Path: toDir}).Error; err != nil {
			return errors.Wrapf(err, "디렉터리 내부 파일 수정 실패 %q", fromDir)
		}

		return nil
	}
}

func (d DirectoryData) ModifyDirectory(fromDir string, toDir string, fromFile string, toFile string) error {
	//tx := d.db.Begin()
	//defer func() {
	//	if err := tx.Rollback().Error; err != nil {
	//		fmt.Println(err)
	//	}
	//}()

	if err := d.db.Transaction(d.modify(fromDir, toDir, fromFile, toFile)); err != nil {
		return errors.Wrap(err, "failed to modify directory")
	}
	//
	//
	//
	//if err := tx.Commit().Error; err != nil {
	//	return errors.Wrap(err, "failed to commit")
	//}

	return nil
}

func (d DirectoryData) LoadDirectoryList(path string) (*DirectoryList, error) {
	panic("implement me")
}

func (d DirectoryData) LoadDirectory(path string) (*model.Directory, error) {
	panic("implement me")
}

func (d *DirectoryData) SetDatabase(db *gorm.DB) {
	d.db = db
}

func (d DirectoryData) GetDatabase() *gorm.DB {
	return d.db
}
