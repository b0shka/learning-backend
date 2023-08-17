package handler

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/b0shka/backend/internal/service"
	mock_service "github.com/b0shka/backend/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_sendCodeEmail(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, email string)

	tests := []struct {
		name         string
		inputBody    string
		email        string
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name:      "ok",
			inputBody: `{"email":"email@ya.ru"}`,
			email:     "email@ya.ru",
			mockBehavior: func(s *mock_service.MockUsers, email string) {
				ctx := context.Background()
				s.EXPECT().SendCodeEmail(ctx, email).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		// {
		// 	name:         "empty fields",
		// 	inputBody:    "",
		// 	mockBehavior: func(s *mock_service.MockUsers, email string) {},
		// 	statusCode:   400,
		// 	responseBody: `{"error":"invalid input body"}`,
		// },
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			users := mock_service.NewMockUsers(c)
			testCase.mockBehavior(users, testCase.email)

			services := &service.Services{Users: users}
			handler := Handler{services: services}

			r := gin.New()
			r.POST("/users/send-code", handler.sendCodeEmail)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(
				"POST",
				"/users/send-code",
				bytes.NewBufferString(testCase.inputBody),
			)

			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, testCase.statusCode)
			assert.Equal(t, w.Body.String(), testCase.responseBody)
		})
	}
}
