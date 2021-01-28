package main

import (
	"strconv"
	"tutorial/model"
	"tutorial/service"

	"github.com/labstack/echo/middleware"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/labstack/echo"
)

type DatabaseInfo struct {
	acct string
	pwd  string
	host string
	port int
	dbn  string
	db   *gorm.DB
}

func (i *DatabaseInfo) ConnectDatabase() error {
	dsn := i.acct + ":" + i.pwd + "@tcp(" + i.host + ":" + strconv.Itoa(i.port) + ")/" + i.dbn + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return err
	}

	i.db = db

	return nil
}

func main() {
	var dbi = DatabaseInfo{
		acct: "root",
		pwd:  "1234",
		host: "127.0.0.1",
		port: 3306,
		dbn:  "tutorial",
		db:   nil,
	}
	err := dbi.ConnectDatabase()

	if err != nil {
		panic("데이터베이스 연결에 실패 하였습니다.")
	}

	// orm 구조체를 통한 자동 테이블 생성
	if err = dbi.db.AutoMigrate(&model.User{}); err != nil {
		return
	}
	if err = dbi.db.AutoMigrate(&model.File{}); err != nil {
		return
	}
	if err = dbi.db.AutoMigrate(&model.Directory{}); err != nil {
		return
	}

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})

	userService := new(service.UserService)
	userService.InitUserService(dbi.db)

	fileService := new(service.FileService)
	fileService.InitFileService(dbi.db)

	dirService := new(service.DirectoryService)
	dirService.InitDirService(dbi.db)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 회원
	e.POST("/user", userService.SaveUser)
	e.POST("/user/login", userService.LoginUser)
	e.POST("/user/logout", userService.LogoutUser)
	e.GET("/user/:id", userService.GetUser)
	e.DELETE("/user/:id", userService.DeleteUser)

	// 파일
	e.POST("/file/upload", fileService.UploadFile, fileService.AuthFileUpload, dirService.CheckDirectory)
	e.PUT("/file/restore", fileService.RestoreFile, fileService.AuthFileRestore)
	e.DELETE("/file", fileService.DeleteFile, fileService.AuthFileDelete)
	e.GET("/file/download/:id", fileService.DownloadFile)
	e.GET("/files/:id", fileService.GetFileInfo)
	e.GET("/files/list", fileService.GetFileList)

	//디렉토리
	e.POST("/dir", dirService.Create)
	e.DELETE("/dir", dirService.Delete)
	e.PUT("/dirs/:id", dirService.Rename)
	e.GET("/dirs/:id", dirService.GetDirInfo)
	e.GET("/dirs", dirService.GetDirList)

	//코멘트

	e.Logger.Fatal(e.Start(":1323"))
}
