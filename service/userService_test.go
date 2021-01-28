package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tutorial/dao"
	"tutorial/mocks"
	"tutorial/model"

	"golang.org/x/crypto/bcrypt"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
)

func TestUserService_SaveUser(t *testing.T) {
	type args struct {
		userJSON string
		mockFunc func(ctrl *gomock.Controller) dao.UserDAO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ERR: user가 정상적인 값이 아닐 때",
			args: args{
				userJSON: `["sangil"]`,
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					return mocks.NewMockUserDAO(nil)
				},
			},
			wantErr: true,
		},
		{
			name: "ERR: bcrypt error",
			args: args{
				userJSON: `{
					"ID": "dbdbtkdgur",
					"Pwd": "1234",
					"status": 1
				}`,
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					monkey.Patch(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
						return nil, fmt.Errorf("test")
					})
					m := mocks.NewMockUserDAO(nil)

					return m
				},
			},
			wantErr: true,
		},
		{
			name: "ERR: user 중복된 이메일일 때",
			args: args{
				userJSON: `{
					"ID": "dbdbtkdgur",
					"Pwd": "1234",
					"status": 1
				}`,
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					monkey.Patch(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
						return []byte("1234"), nil
					})
					m := mocks.NewMockUserDAO(ctrl)
					user := &model.User{ID: "dbdbtkdgur", Pwd: "1234", Status: 1}
					m.EXPECT().Post(user).Return(fmt.Errorf("duplicated")).AnyTimes()

					return m
				},
			},
			wantErr: true,
		},
		{
			name: "OK",
			args: args{
				userJSON: `{
					"ID": "dbdbtkdgur",
					"Pwd": "1234",
					"status": 1
				}`,
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					monkey.Patch(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
						return []byte("1234"), nil
					})
					m := mocks.NewMockUserDAO(ctrl)
					user := &model.User{ID: "dbdbtkdgur", Pwd: "1234", Status: 1}
					m.EXPECT().Post(user).Return(nil).AnyTimes()

					return m
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(tt.args.userJSON))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			res := httptest.NewRecorder()
			ctx := e.NewContext(req, res)

			userDAO := tt.args.mockFunc(ctrl)
			u := UserService{
				memberDAO: userDAO,
			}

			if err := u.SaveUser(ctx); (err != nil) != tt.wantErr {
				t.Errorf("SaveUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	type args struct {
		userID   string
		mockFunc func(ctrl *gomock.Controller) dao.UserDAO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// 가상 테스트 시나리오 ( 실제 동작과 차이가 있을 수 있음 )
		{
			name: "ERR: test는 아이디로 사용 할 수 없음.",
			args: args{
				userID: "test",
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					m := mocks.NewMockUserDAO(ctrl)
					m.EXPECT().Get("test").Return(nil, fmt.Errorf("유저 아이디로 test 는 사용 할 수 없음")).AnyTimes()
					return m
				},
			},
			wantErr: true,
		},
		{
			name: "OK",
			args: args{
				userID: "ysh9579",
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					m := mocks.NewMockUserDAO(ctrl)
					m.EXPECT().Get("ysh9579").Return(&model.User{ID: "ysh9579", Pwd: "1234", Status: 1}, nil).AnyTimes()
					return m
				},
			},

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/user", nil)
			res := httptest.NewRecorder()
			ctx := e.NewContext(req, res)

			ctx.SetParamNames("id")
			ctx.SetParamValues(tt.args.userID)

			userDAO := tt.args.mockFunc(ctrl)
			u := UserService{
				memberDAO: userDAO,
			}

			if err := u.GetUser(ctx); (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	type args struct {
		jsonData string
		mockFunc func(ctrl *gomock.Controller) dao.UserDAO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ERR: Password is empty",
			args: args{
				jsonData: `{"ID":"ysh9579", "Pwd":""}`,
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					monkey.Patch(model.User.CheckPasswordHash, func(user model.User, pwd string) bool {
						return false
					})

					m := mocks.NewMockUserDAO(ctrl)
					m.EXPECT().Get("ysh9579").Return(&model.User{ID: "ysh9579", Pwd: "", Status: 1}, fmt.Errorf("패스워드가 잘못 저장 되었습니다. \n")).AnyTimes()

					return m
				},
			},
			wantErr: true,
		},
		{
			name: "OK",
			args: args{
				jsonData: `{"ID":"ysh9579", "Pwd":"1234"}`,
				mockFunc: func(ctrl *gomock.Controller) dao.UserDAO {
					monkey.Patch(model.User.CheckPasswordHash, func(user model.User, pwd string) bool {
						return true
					})

					m := mocks.NewMockUserDAO(ctrl)
					m.EXPECT().Get("ysh9579").Return(&model.User{ID: "ysh9579", Pwd: "1234", Status: 1}, nil).AnyTimes()

					return m
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserDAO := tt.args.mockFunc(ctrl)

			u := UserService{memberDAO: mockUserDAO}

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(tt.args.jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			res := httptest.NewRecorder()

			ctx := e.NewContext(req, res)

			if err := u.LoginUser(ctx); (err != nil) != tt.wantErr {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
