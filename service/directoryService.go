package service

import (
	"fmt"
	"net/http"
	"os"
	"tutorial/dao"

	"github.com/pkg/errors"

	"github.com/labstack/echo"

	"gorm.io/gorm"
)

const RootDirPath = "/Users/sanghyuk/Desktop/TestCloudBox"
const TrashDirPath = "/Users/sanghyuk/Desktop/Trash"

type Path string

func (p Path) GetPathWithRoot() string {
	return RootDirPath + string(p)
}

func (p Path) GetPathWithTrash() string {
	return TrashDirPath + string(p)
}

func (p Path) String() string {
	return string(p)
}

type UploadService struct {
	fileService FileService
	dirService  DirectoryService
}

type DirectoryService struct {
	DirDAO dao.DirectoryDAO
}

func (d *DirectoryService) InitDirService(db *gorm.DB) {
	if d.DirDAO == nil {
		d.DirDAO = new(dao.DirectoryData)
	}

	d.DirDAO.SetDatabase(db)
}

func (d DirectoryService) CheckDirectory(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("CheckDirectory 미들웨어 시작...")

		path := Path(c.FormValue("path"))

		if err := d.CreateDir(path.GetPathWithRoot()); err != nil {
			return err
		}

		err := next(c)

		fmt.Println("CheckDirectory 미들웨어 종료..")

		return err
	}
}

func (d DirectoryService) Create(c echo.Context) error {
	path := Path(c.FormValue("path"))

	if err := d.CreateDir(path.GetPathWithRoot()); err != nil {
		return c.String(http.StatusBadRequest, "디렉터리 생성 실패 path : "+path.String())
	}

	return c.String(http.StatusOK, "디렉터리 생성 성공 path : "+path.String())
}

func (d DirectoryService) CreateDir(path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 디렉토리가 존재 하지 않는다면 디렉터리 생성 이후 디비에도 저장
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "디렉터리 생성 실패 %q", path)
		}
		// 디비에 디렉터리 저장
		err = d.DirDAO.CreateDirectory(path)

		if err != nil {
			return errors.Wrapf(err, "디렉터리 정보 디비 저장 실패 %q", path)
		}
	}

	return nil
}

func (d DirectoryService) Delete(c echo.Context) error {

	path := Path(c.FormValue("path"))

	oldPath := path.GetPathWithRoot()
	deletePath := path.GetPathWithTrash()

	if err := os.Rename(oldPath, deletePath); err != nil {
		return c.String(http.StatusBadRequest, "디렉터리 휴지통으로 이동 실패")
	}

	// 1. 디렉토리 조회
	// 2. 디렉토리 삭제
	// 3. 디렉토리 내부의 파일 삭제.

	if err := d.DirDAO.DeleteDirectory(path.GetPathWithRoot(), path.String()); err != nil {
		return c.String(http.StatusBadRequest, "디렉터리 데이터베이스 삭제 실패")
	}

	return c.String(http.StatusOK, "디렉터리 삭제 성공")
}

func (d DirectoryService) Rename(c echo.Context) error {
	oldPath := Path(c.FormValue("path"))
	newPath := Path(c.FormValue("renamePath"))

	if err := os.Rename(oldPath.GetPathWithRoot(), newPath.GetPathWithRoot()); err != nil {
		return c.String(http.StatusBadRequest, "디렉터리 이동 실패 oldPath = "+oldPath.String()+", newPath = "+newPath.String())
	}

	if err := d.DirDAO.ModifyDirectory(oldPath.GetPathWithRoot(), newPath.GetPathWithRoot(), oldPath.String(), newPath.String()); err != nil {
		return c.String(http.StatusBadRequest, "디렉터리 데이터베이스 이동 실패 oldPath = "+oldPath.String()+", newPath = "+newPath.String())
	}

	return c.String(http.StatusOK, "디렉터리 수정 성공")
}

func (d DirectoryService) GetDirInfo(c echo.Context) error {
	return nil
}

func (d DirectoryService) GetDirList(c echo.Context) error {
	return nil
}
