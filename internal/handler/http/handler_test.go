package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/b0shka/backend/internal/handler/http"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	h := handler.NewHandler(&service.Services{}, &auth.Manager{})

	require.IsType(t, &handler.Handler{}, h)
}

func TestNewHandler_InitRoutes(t *testing.T) {
	h := handler.NewHandler(&service.Services{}, &auth.Manager{})
	router := h.InitRoutes()

	ts := httptest.NewServer(router)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ping")
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, http.StatusOK, res.StatusCode)
}
