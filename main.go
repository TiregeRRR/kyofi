package main

import "github.com/TiregeRRR/kyofi/ui"

func main() {
	app, err := ui.New()
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}

	// s, err := ssh.New()
	// if err != nil {
	// 	panic(err)
	// }

	// f, err := s.Open("")
	// if err != nil {
	// 	panic(err)
	// }

	// for i := range f {
	// 	fmt.Printf("%v\n", f[i])
	// }

	// err = s.Copy("local")
	// if err != nil {
	// 	panic(err)
	// }

	// p, err := s.PasteReader()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(p)

	// for p.Next() {
	// 	p.File()
	// }
}
