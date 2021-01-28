package dao

import (
	"database/sql"
	"tutorial/model"

	"github.com/pkg/errors"

	"gorm.io/gorm"
)

type FileDAO interface {
	SaveFile(file *model.File, root string) error
	GetFile(file *model.File) (RowsAffected int64, err error)
	GetFileByID(file *model.File) *gorm.DB
	GetFileListByPath(path string) (*model.FileList, error)
	GetDeleteFile(file *model.File) *gorm.DB
	DeleteFile(file *model.File) error
	RestoreFile(file *model.File) error
	SetDatabase(db *gorm.DB)
	GetDatabase() *gorm.DB
}

type FileData struct {
	db *gorm.DB
}

func (d FileData) SaveFile(file *model.File, root string) error {
	// 디렉토리 아이디 가져오기
	var dir model.Directory

	if err := d.db.First(&dir, "path = ?", file.GetDirPath(root)).Error; err != nil {
		return errors.Wrapf(err, "Fail to Query Get Directory")
	}

	file.DirectoryID = dir.ID

	return d.db.Create(file).Error
}

func (d FileData) GetFile(file *model.File) (RowsAffected int64, err error) {
	tx := d.db.First(file, "file_name = ? AND path = ? AND extension = ?", file.FileName, file.Path, file.Extension)

	return tx.RowsAffected, tx.Error
}

func (d FileData) GetFileByID(file *model.File) *gorm.DB {
	return d.db.Find(file, file.ID)
}

func (d FileData) GetFileListByPath(path string) (*model.FileList, error) {
	fileList := make(model.FileList, 100)

	if err := d.db.Where("path= ?", path).Find(&fileList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fileList, errors.Wrapf(err, "파일을 찾을 수 없습니다.")
		}

		return &fileList, err
	}

	return &fileList, nil
}

func (d FileData) GetDeleteFile(file *model.File) *gorm.DB {
	return d.db.Unscoped().Where("file_name = ? AND extension = ? AND path = ?", file.FileName, file.Extension, file.Path).Find(&file)
}

func (d *FileData) SetDatabase(db *gorm.DB) {
	d.db = db
}

func (d FileData) GetDatabase() *gorm.DB {
	return d.db
}

func (d FileData) DeleteFile(file *model.File) error {
	if err := d.db.Delete(file).Error; err != nil {

		return errors.Wrapf(err, "디비에서 파일 삭제 실패 %q", file.String())
	}

	return nil
}

func (d FileData) RestoreFile(file *model.File) error {
	tx := d.db.Unscoped().Where("file_name = ? AND path = ? AND extension = ?", file.FileName, file.Path, file.Extension).Find(file)

	if tx.RowsAffected == 1 {
		// soft delete 시 채워지는 DeletedAt null 처리
		file.DeletedAt = gorm.DeletedAt(sql.NullTime{Valid: false})

		if err := d.db.Save(file).Error; err != nil {
			return errors.Wrapf(err, "fail to restore file %q", file.String())
		}
	} else if tx.RowsAffected > 1 {
		return errors.Wrapf(tx.Error, "has too many row")
	}

	return nil
}
