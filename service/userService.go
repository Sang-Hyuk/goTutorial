package service

import (
	"fmt"
	"net/http"
	"time"
	"tutorial/dao"
	"tutorial/model"

	"gorm.io/gorm"

	"github.com/labstack/echo"
)

const (
	CookieUserid = "UserID"
)

type UserService struct {
	memberDAO dao.UserDAO
}

func (u *UserService) InitUserService(db *gorm.DB) {
	if u.memberDAO == nil {
		u.memberDAO = new(dao.UserData)
	}

	u.memberDAO.SetDatabase(db)
}

func (u UserService) LoginUser(context echo.Context) error {
	user := new(model.User)

	if err := context.Bind(user); err != nil {
		return err
	}

	loginUser, err := u.memberDAO.Get(user.ID)

	if err != nil {
		return err
	}

	match := loginUser.CheckPasswordHash(user.Pwd)

	if match {
		// 로그인 성공
		cookie, _ := context.Cookie(CookieUserid)
		if cookie == nil {
			cookie := new(http.Cookie)
			cookie.Name = CookieUserid
			cookie.Value = loginUser.ID
			cookie.Expires = time.Now().Add(24 * time.Hour)
			cookie.HttpOnly = true
			context.SetCookie(cookie)
		} else {
			cookie.Value = loginUser.ID
			cookie.Expires = time.Now().Add(24 * time.Hour)
			context.SetCookie(cookie)
		}

		return context.String(http.StatusOK, "로그인 성공: "+loginUser.String())
	} else {
		// 로그인 실패
		return context.String(http.StatusNoContent, "로그인 실패: "+loginUser.String())
	}
}

func (u UserService) LogoutUser(context echo.Context) error {
	cookie, _ := context.Cookie(CookieUserid)

	if cookie != nil {
		cookie.MaxAge = -1
		context.SetCookie(cookie)
		return context.String(http.StatusOK, "로그아웃 성공: "+cookie.Value)
	}

	return context.String(http.StatusOK, "이미 로그아웃 되셨습니다.")
}

func (u UserService) SaveUser(context echo.Context) error {
	user := new(model.User)

	if err := context.Bind(user); err != nil {
		return err
	}

	hash, err := user.HashPassword()

	if err != nil {
		return err
	}

	user.Pwd = hash

	if err := u.memberDAO.Post(user); err != nil {
		return err
	}

	return context.String(http.StatusCreated, "회원 등록 완료 : "+user.String())
}

func (u UserService) GetUser(c echo.Context) error {
	id := c.Param("id")

	findUser, err := u.memberDAO.Get(id)

	if err != nil {
		fmt.Printf("get error: %v\n", err)
		return err
	}

	return c.String(http.StatusOK, "[찾은 회원 정보] : "+findUser.String())
}

func (u UserService) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	if err := u.memberDAO.Delete(id); err != nil {
		return err
	}

	return c.String(http.StatusOK, "["+id+"] 유저가 삭제 되었습니다.")
}
