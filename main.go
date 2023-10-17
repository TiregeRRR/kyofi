package main

import (
	"github.com/TiregeRRR/kyofi/ui"
)

func main() {
	app, err := ui.New()
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
