package foo

import (
	"testing"
	"tutorial/mocks"

	"github.com/golang/mock/gomock"
)

func TestFoo(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	m := mocks.NewMockFoo(ctrl)

	m.EXPECT().Do(gomock.Eq(99)).Return(101).AnyTimes()

	Bar(m)
}
