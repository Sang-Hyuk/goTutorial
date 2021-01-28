package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"tutorial/dao"
	"tutorial/mocks"
	"tutorial/model"

	"github.com/pkg/errors"

	"bou.ke/monkey"

	"github.com/golang/mock/gomock"

	"github.com/labstack/echo"
)

var mockHandler = func(c echo.Context) error { return nil }

func GetTestFileInfo() (os.FileInfo, error) {
	file, err := os.Open("/Users/sanghyuk/Desktop/TestSendfile/Test.txt")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fi, err := file.Stat()

	if err != nil {
		return nil, err
	}

	return fi, err
}

func TestFileService_AuthFileUpload(t *testing.T) {
	type fields struct {
		FileDAO func(ctrl *gomock.Controller) dao.FileDAO
	}
	type args struct {
		filename  string
		extension string
		path      string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ERROR: Vaildate false empty fileName",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				return nil
			}},
			args: args{"", "txt", "/path"},
			want: true,
		},
		{
			name: "ERROR: Validate false empty extension",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				return nil
			}},
			args: args{"test1", "", "/path"},
			want: true,
		},
		{
			name: "ERROR: Validate false empty path",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				return nil
			}},
			args: args{"test1", "txt", ""},
			want: true,
		},
		{
			name: "ERROR: Duplicate File",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				m := mocks.NewMockFileDAO(ctrl)
				file := model.File{
					FileName:    "test1",
					Extension:   "txt",
					Path:        "/Users",
					DirectoryID: 0,
				}
				m.EXPECT().GetFile(&file).Return(int64(1), fmt.Errorf("DuplFile Name %q.%q", file.FileName, file.Extension)).AnyTimes()

				return m
			}},
			args: args{"test1", "txt", "/Users"},
			want: true,
		},
		{
			name: "OK",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				m := mocks.NewMockFileDAO(ctrl)
				file := model.File{
					FileName:    "test1",
					Extension:   "txt",
					Path:        "/Users",
					DirectoryID: 0,
				}
				m.EXPECT().GetFile(&file).Return(int64(0), nil).AnyTimes()

				return m
			}},
			args: args{"test1", "txt", "/Users"},
			want: false,
		},
	}

	fi, err := GetTestFileInfo()

	if err != nil {
		fmt.Println("테스트 파일 정보 가져오기 실패")
		return
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			writer.WriteField("filename", tt.args.filename)
			writer.WriteField("extension", tt.args.extension)
			writer.WriteField("path", tt.args.path)

			part, err := writer.CreateFormFile("file", fi.Name())

			if err != nil {
				t.Errorf("Fail to create form file\n")
				return
			}

			part.Write([]byte("Test"))

			if err := writer.Close(); err != nil {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/file", body)
			req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
			res := httptest.NewRecorder()

			e := echo.New()
			ctx := e.NewContext(req, res)

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			f := FileService{
				FileDAO: tt.fields.FileDAO(ctl),
			}

			if got := f.AuthFileUpload(mockHandler)(ctx); (got != nil) != tt.want {
				t.Errorf("AuthFileUpload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileService_UploadFile(t *testing.T) {
	type fields struct {
		FileDAO func(ctrl *gomock.Controller) dao.FileDAO
	}
	type args struct {
		filename     string
		extension    string
		path         string
		bVaidateFile bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ERROR: File is not validate",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				return nil
			}},
			args: args{
				filename:     "Test",
				extension:    "txt",
				path:         "/",
				bVaidateFile: false,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Fail to open file",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				monkey.Patch(OpenFile, func(fh *multipart.FileHeader) (multipart.File, error) {
					return nil, errors.Errorf("fail to open file \n")
				})

				return nil
			}},
			args: args{
				filename:     "Test",
				extension:    "txt",
				path:         "/",
				bVaidateFile: true,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Fail to create file",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				monkey.Patch(CreateFile, func(name string) (*os.File, error) {
					return nil, errors.Errorf("fail to create file \n")
				})

				return nil
			}},
			args: args{
				filename:     "Test",
				extension:    "txt",
				path:         "/",
				bVaidateFile: true,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Fail to copy file",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				monkey.Patch(CopyFile, func(dst io.Writer, src io.Reader) error {
					return errors.Errorf("fail to copy file\n")
				})

				return nil
			}},
			args: args{
				filename:     "Test",
				extension:    "txt",
				path:         "",
				bVaidateFile: true,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Fail to save file",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				m := mocks.NewMockFileDAO(ctrl)
				file := model.File{
					FileName:    "Test",
					Extension:   "txt",
					Path:        "",
					DirectoryID: 0,
				}
				m.EXPECT().SaveFile(&file, RootDirPath).Return(errors.Errorf("fail to save file \n")).AnyTimes()
				return m
			}},
			args: args{
				filename:     "Test",
				extension:    "txt",
				path:         "",
				bVaidateFile: true,
			},
			wantErr: true,
		},
		{
			name: "OK",
			fields: fields{func(ctrl *gomock.Controller) dao.FileDAO {
				m := mocks.NewMockFileDAO(ctrl)
				file := model.File{
					FileName:    "Test",
					Extension:   "txt",
					Path:        "",
					DirectoryID: 0,
				}
				m.EXPECT().SaveFile(&file, RootDirPath).Return(nil).AnyTimes()
				return m
			}},
			args: args{
				filename:     "Test",
				extension:    "txt",
				path:         "",
				bVaidateFile: true,
			},
			wantErr: false,
		},
	}

	fi, err := GetTestFileInfo()

	if err != nil {
		fmt.Println("테스트 파일 정보 가져오기 실패")
		return
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			writer.WriteField("filename", tt.args.filename)
			writer.WriteField("extension", tt.args.extension)
			writer.WriteField("path", tt.args.path)

			if tt.args.bVaidateFile {
				part, err := writer.CreateFormFile("file", fi.Name())

				if err != nil {
					t.Errorf("fail to create form file \n")
					return
				}

				part.Write([]byte("Test"))

			}

			if err := writer.Close(); err != nil {
				t.Errorf("fail to close writer \n")
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/file", body)
			req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
			res := httptest.NewRecorder()

			e := echo.New()
			ctx := e.NewContext(req, res)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			monkey.UnpatchAll()

			f := FileService{
				FileDAO: tt.fields.FileDAO(ctrl),
			}
			if err := f.UploadFile(ctx); (err != nil) != tt.wantErr {
				t.Errorf("UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
