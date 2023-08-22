package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/b0shka/backend/internal/config"
	handler "github.com/b0shka/backend/internal/handler/http"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	h := handler.NewHandler(&service.Services{}, &auth.PasetoManager{})

	require.IsType(t, &handler.Handler{}, h)
}

func TestNewHandler_InitRoutes(t *testing.T) {
	h := handler.NewHandler(&service.Services{}, &auth.PasetoManager{})
	router := h.InitRoutes(&config.Config{})

	ts := httptest.NewServer(router)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ping")
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, http.StatusOK, res.StatusCode)
}
