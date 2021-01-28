package dao

import (
	"testing"
	"tutorial/model"

	"github.com/pkg/errors"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"

	"gorm.io/gorm"
)

func TestFileData_SaveFile(t *testing.T) {
	const RootDirPath = "/Users/sanghyuk/Desktop/TestCloudBox"

	type args struct {
		file *model.File
		root string
		mock func(mock sqlmock.Sqlmock)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			// ERR: file이 nil일 때
			// ERR: 유효하지 않은 root
			name: "ERROR: Fail to Query Get Directory",
			args: args{
				file: &model.File{Path: "/UndefinedPath"},
				root: RootDirPath,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery("SELECT").
						WithArgs(RootDirPath + "/UndefinedPath").
						WillReturnError(errors.New("존재하지 않는 디렉터리 입니다."))
				},
			},
			wantErr: true,
		},
		{
			// ERR: 저장실패
			name: "ERROR: fail to create directory",
			args: args{
				file: &model.File{Path: "/Exist"},
				root: RootDirPath,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
					mock.ExpectExec("INSERT").WillReturnError(errors.New("파일 생성 실패"))
				},
			},
			wantErr: true,
		},
		{
			// ERR: 저장실패
			name: "ERROR: fail to create directory",
			args: args{
				file: &model.File{Path: "/Exist"},
				root: RootDirPath,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
					mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(10, 1))
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Error(err)
			}
			gdb, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{})

			defer db.Close()

			mocker := tt.args.mock
			if mocker != nil {
				mocker(mock)
			}

			d := FileData{
				db: gdb.Debug(),
			}

			if err := d.SaveFile(tt.args.file, tt.args.root); (err != nil) != tt.wantErr {
				t.Errorf("SaveFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there wewe unfulfilled expectations : %s\n", err)
			}

		})
	}
}
