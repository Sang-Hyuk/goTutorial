package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"tutorial/dao"
	"tutorial/model"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type FileService struct {
	FileDAO dao.FileDAO
}

func (f *FileService) InitFileService(db *gorm.DB) {
	if f.FileDAO == nil {
		f.FileDAO = new(dao.FileData)
	}

	f.FileDAO.SetDatabase(db)
}

func OpenFile(file *multipart.FileHeader) (src multipart.File, err error) {
	fmt.Println("OpenFile Start...")
	src, err = file.Open()
	fmt.Println("Openfile End...")
	return src, err
}

func CreateFile(name string) (*os.File, error) {
	return os.Create(name)
}

func CopyFile(dst io.Writer, src io.Reader) error {
	if _, err := io.Copy(dst, src); err != nil {
		return errors.Wrapf(err, "fail to Copy file src : %q, dst : %q\n", src, dst.(*os.File).Name())
	}
	return nil
}

func (f FileService) UploadFile(c echo.Context) error {
	fmt.Println("UploadFile 시작...")

	fileModel := model.File{FileName: c.FormValue("filename"), Extension: c.FormValue("extension"), Path: c.FormValue("path")}

	file, err := c.FormFile("file")

	if err != nil {
		return errors.Wrapf(err, "fail to get file name\n")
	}

	src, err := OpenFile(file)

	if src != nil {
		defer src.Close()
	}

	if err != nil {
		return errors.Wrapf(err, "fail to open file \n")
	}

	dstFilepath := fileModel.GetFullPath(RootDirPath)

	dst, err := CreateFile(dstFilepath)

	if err != nil {
		return errors.Wrapf(err, "fail to Create file name : %q\n", fileModel.FileName)
	}

	if dst != nil {
		defer dst.Close()
	}

	if err := CopyFile(dst, src); err != nil {
		return errors.Wrapf(err, "fail to Copy file src : %q, dst : %q\n", src, dst.Name())
	}

	if err := f.FileDAO.SaveFile(&fileModel, RootDirPath); err != nil {
		return errors.Wrapf(err, "Fail to Save File %q\n", fileModel.String())
	}

	fmt.Println("UploadFile 종료...")
	return c.String(http.StatusOK, "파일 업로드 성공 : "+fileModel.String())
}

func (f FileService) RestoreFile(c echo.Context) error {
	fmt.Println("RestoreFile 시작...")

	fileModel := model.File{FileName: c.FormValue("filename"), Extension: c.FormValue("extension"), Path: c.FormValue("path")}

	// 실제 파일 복원을 위해 준비해야함
	if err := f.FileDAO.GetDeleteFile(&fileModel).Error; err != nil {
		return errors.Wrapf(err, "삭제된 파일 조회 실패 %q", fileModel.String())
	}

	restorePath := fileModel.GetFullPath(RootDirPath)
	deletePath := TrashDirPath + fileModel.GetFileNameWithExt(true)

	if err := os.Rename(deletePath, restorePath); err != nil {
		return errors.Wrapf(err, "휴지통에서 복원 디렉터리로 이동실패 RestorePath : %q", fileModel.Path)
	}

	if err := f.FileDAO.RestoreFile(&fileModel); err != nil {
		return errors.Wrapf(err, "데이터 베이스 파일 복원 실패 %q", fileModel.String())
	}

	fmt.Println("RestoreFile 종료...")

	return c.String(http.StatusOK, "파일 복원 성공 : "+fileModel.String())
}

func (f FileService) DeleteFile(c echo.Context) error {
	fmt.Println("DeleteUser 시작...")

	fileModel := model.File{FileName: c.FormValue("filename"), Extension: c.FormValue("extension"), Path: c.FormValue("path")}

	if _, err := f.FileDAO.GetFile(&fileModel); err != nil {
		return errors.Wrapf(err, "삭제 할 파일 정보 조회 실패 %q", fileModel.String())
	}

	oldPath := fileModel.GetFullPath(RootDirPath)
	deletePath := TrashDirPath + fileModel.GetFileNameWithExt(true)

	if err := os.Rename(oldPath, deletePath); err != nil {
		return errors.Wrapf(err, "파일 휴지통으로 이동 실패 %q", fileModel.String())
	}

	if err := f.FileDAO.DeleteFile(&fileModel); err != nil {
		return err
	}

	fmt.Println("DeleteUser 종료...")

	return c.String(http.StatusOK, "파일 삭제 성공 : "+fileModel.String())
}

func (f FileService) AuthFileUpload(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("AuthFileUpload 미들웨어 시작...")

		fileModel := model.File{FileName: c.FormValue("filename"), Extension: c.FormValue("extension"), Path: c.FormValue("path")}

		bResult, errorCode := fileModel.CheckValidate()

		if !bResult {
			return errors.Errorf("파일 업로드 실패 : %q \n", fileModel.GetErrorMsg(errorCode))
		}

		bResult, errorCode = f.CheckDuplication(&fileModel)

		if bResult {
			return errors.Errorf("파일 업로드 실패 : %q \n", fileModel.GetErrorMsg(errorCode))
		}

		err := next(c)

		fmt.Println("AuthFileUpload 미들웨어 종료...")

		return err
	}
}

func (f FileService) AuthFileDelete(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("AuthFileDelete 미들웨어 시작...")

		fileModel := model.File{FileName: c.FormValue("filename"), Extension: c.FormValue("extension"), Path: c.FormValue("path")}

		bResult, _ := f.CheckDuplication(&fileModel)

		if !bResult {
			return c.String(http.StatusOK, "삭제할 파일이 디비에 존재 하지 않습니다")
		}

		fi, err := os.Stat(fileModel.GetFullPath(RootDirPath))

		if err != nil {
			return c.String(http.StatusOK, "삭제할 파일이 경로에 존재 하지 않습니다")
		}

		if fi.IsDir() {
			return c.String(http.StatusOK, "디렉토리는 삭제 할 수 없습니다.")
		}

		err = next(c)

		fmt.Println("AuthFileDelete 미들웨어 종료...")

		return err
	}
}

func (f FileService) AuthFileRestore(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("authFileRestore 미들웨어 시작...")

		fileModel := model.File{FileName: c.FormValue("filename"), Extension: c.FormValue("extension"), Path: c.FormValue("path")}

		tx := f.FileDAO.GetDeleteFile(&fileModel)

		if tx.RowsAffected <= 0 {
			return errors.New("복원할 파일이 디비에 존재 하지 않습니다")
		}

		deletePath := TrashDirPath + fileModel.GetFileNameWithExt(true)

		_, err := os.Stat(deletePath)

		if err != nil {
			return errors.New("복원할 파일이 휴지통에 존재 하지 않습니다")
		}

		err = next(c)

		fmt.Println("authFileRestore 미들웨어 종료...")

		return err
	}
}

func (f FileService) CheckDuplication(file *model.File) (bool, model.FileErrorCode) {
	RowsAffected, _ := f.FileDAO.GetFile(file)

	if RowsAffected > 0 {
		return true, model.ErrorDuplFileName
	}

	return false, model.ErrorNone
}

func (f FileService) GetFileInfo(c echo.Context) error {

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		return errors.Wrapf(err, "파일 ID의 형식이 잘못 되었습니다.")
	}

	file := model.File{
		Model: gorm.Model{ID: uint(fileID)},
	}

	tx := f.FileDAO.GetFileByID(&file)

	if tx.RowsAffected <= 0 {
		return errors.Wrapf(tx.Error, "존재하지않는 파일입니다.")
	}

	return c.String(http.StatusOK, "[파일정보] "+file.String())
}

func (f FileService) GetFileList(c echo.Context) error {
	fileList, err := f.FileDAO.GetFileListByPath(c.QueryParam("path"))

	if err != nil {
		return errors.Wrapf(err, "파일리스트 조회 실패")
	}

	return c.String(http.StatusOK, "[파일정보] "+fileList.String())
}

func (f FileService) DownloadFile(c echo.Context) error {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		return errors.Wrapf(err, "파일 ID의 형식이 잘못 되었습니다.")
	}

	file := model.File{
		Model: gorm.Model{ID: uint(fileID)},
	}

	tx := f.FileDAO.GetFileByID(&file)

	if tx.RowsAffected <= 0 {
		return errors.Wrapf(tx.Error, "존재하지않는 파일입니다.")
	}

	return c.Attachment(file.GetFullPath(RootDirPath), "test1.pdf")
}
