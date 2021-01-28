package dao

import (
	"testing"
	"tutorial/mocks"
	"tutorial/model"

	"github.com/golang/mock/gomock"
)

func TestUserDAO(t *testing.T) {
	ctrl := gomock.NewController(t)

	// go1.14+ 부터는 테스트 객체를 전달하는 경우 호출할 필요가 없다? 자동으로 생김
	defer ctrl.Finish()

	m := mocks.NewMockUserDAO(ctrl)

	user := &model.User{ID: "admin", Pwd: "1234", Status: 1}

	m.EXPECT().Post(user).Return(nil).AnyTimes()
	m.EXPECT().Get(user.ID).Return(&model.User{ID: "admin", Pwd: "1234", Status: 1}, nil).AnyTimes()
	m.EXPECT().Delete(user.ID).Return(nil).AnyTimes()

	//gomock.InOrder(
	//	m.EXPECT().Delete(umMachedID).Return(fmt.Errorf("err....")).AnyTimes(),
	//	m.EXPECT().Delete(userID).Return(nil).AnyTimes(),
	//)

	//if err := m.Delete(userID); err != nil {
	//	t.Errorf("Delete(%q) = nil", userID)
	//}

	DeleteAfterTester(m)
}
