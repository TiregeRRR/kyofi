package main

import (
	"os"

	"github.com/TiregeRRR/kyofi/modules/file"
	"github.com/TiregeRRR/kyofi/ui"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	left := file.New(path)
	right := file.New(path)

	app, err := ui.New(left, right)
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
