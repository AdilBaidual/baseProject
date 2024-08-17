package main

import (
	"github.com/AdilBaidual/baseProject/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.NewApp()).Run()
}
