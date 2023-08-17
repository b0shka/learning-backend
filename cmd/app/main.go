package main

import (
	"github.com/b0shka/backend/internal/app"
)

const configPath = "configs"

func main() {
	app.Run(configPath)
}
