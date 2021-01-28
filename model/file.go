package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	FileName    string `form:"filename"`
	Extension   string `form:"extension"`
	Path        string `form:"path"`
	DirectoryID uint
}

type FileList []File

type FileErrorCode int

const (
	ErrorNone = iota
	ErrorEmptyFileName
	ErrorEmptyFileExtention
	ErrorEmptyFilePath
	ErrorDuplFileName
)

func (f File) CheckValidate() (bool, FileErrorCode) {
	if f.FileName == "" {
		return false, ErrorEmptyFileName
	} else if f.Extension == "" {
		return false, ErrorEmptyFileExtention
	} else if f.Path == "" {
		return false, ErrorEmptyFilePath
	}

	return true, ErrorNone
}

//func (f File) CheckValidate() (error)  {
//	switch  {
//	case f.FileName == "":
//		return fmt.Errorf("파일명이 누락 되었습니다.")
//	case f.Extension == "" :
//		return fmt.Errorf("확장자가 누락 되었습니다.")
//	case f.Path == "":
//		return fmt.Errorf("파일 저장 경로가 누락 되었습니다.")
//	default:
//		return nil
//	}
//}

func (f File) GetErrorMsg(code FileErrorCode) string {
	switch code {
	case ErrorEmptyFileName:
		return "파일명이 누락 되었습니다."
	case ErrorEmptyFileExtention:
		return "확장자가 누락 되었습니다."
	case ErrorEmptyFilePath:
		return "파일 저장 경로가 누락 되었습니다."
	case ErrorDuplFileName:
		return "파일명이 중복 되었습니다."
	default:
		return ""
	}

	return ""
}

func (f File) GetFileNameWithExt(bPreSlash bool) string {
	if bPreSlash && !strings.Contains(f.FileName, "/") {
		return "/" + f.FileName + "." + f.Extension
	} else {
		return f.FileName + "." + f.Extension
	}
}

func (f File) GetFullPath(root string) string {
	return root + f.Path + f.GetFileNameWithExt(true)
}

func (f File) GetDirPath(root string) string {
	return root + f.Path
}

func (f File) String() string {
	return fmt.Sprintf("[ID] : %d, [fileName] : %s, [extention] : %s, [path] : %s ", f.ID, f.FileName, f.Extension, f.Path)
}

func (l FileList) String() string {
	var strReturn string

	for _, value := range l {
		strReturn += value.String() + "\n"
	}

	return fmt.Sprintf(strReturn)
}
