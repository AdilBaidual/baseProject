package main

import (
	"go.uber.org/fx"

	"github.com/AdilBaidual/baseProject/internal/app"
)

func main() {
	fx.New(app.NewApp()).Run()
}
