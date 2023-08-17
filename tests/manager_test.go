package tests

import (
	"fmt"
	"testing"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	testTable := []string{
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTQ3ODU0NDcsInN1YiI6IjY0ZGNjYzNiMjdmOTdiZDA2OTBmMTMyMCJ9.9qgrh2-_ZYXD0S7zXq1zkUtTvEycYr9Fb4ZJVwXRokQ",
		"",
	}

	const configPath = "configs"
	cfg, err := config.InitConfig(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	tokenManager, err := auth.NewManager(cfg.Auth.SecretKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, testCase := range testTable {
		id, err := tokenManager.Parse(testCase)

		assert.NoError(t, err)
		assert.NotEmpty(t, id)
	}
}
