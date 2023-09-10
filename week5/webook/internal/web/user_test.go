package web

import (
	"bytes"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	svcmocks "gitee.com/geekbang/basic-go/webook/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestEmailPattern 用来验证我们的邮箱正则表达式对不对
func TestEmailPattern(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool
	}{
		{
			name:  "不带@",
			email: "123456",
			match: false,
		},
		{
			name:  "带@ 但是没后缀",
			email: "123456@",
			match: false,
		},
		{
			name:  "合法邮箱",
			email: "123456@qq.com",
			match: true,
		},
	}

	h := NewUserHandler(nil, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := h.emailRegexExp.MatchString(tc.email)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestPasswordPattern(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		match    bool
	}{
		{
			name:     "合法密码",
			password: "Hello#world123",
			match:    true,
		},
		{
			name:     "没有数字",
			password: "Hello#world",
			match:    false,
		},
		{
			name:     "没有特殊字符",
			password: "Helloworld123",
			match:    false,
		},
		{
			name:     "长度不足",
			password: "he!123",
			match:    false,
		},
	}

	h := NewUserHandler(nil, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := h.passwordRegexExp.MatchString(tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestLoginJWT(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqbody  string
		wantcode int
		wantBody string
	}{
		{
			name: "登陆成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return usersvc
			},
			reqbody: `{"email":"123@qq.com","password":"1234567!@#"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest("POST", "/users/login",
				bytes.NewBuffer([]byte(tc.reqbody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			//t.Log(resp)
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantcode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body)
		})
	}

}
