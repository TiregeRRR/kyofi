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

	// s, err := minio.New(minio.Opts{
	// 	Endpoint:        "172.18.0.2:9000",
	// 	AccessKeyID:     "root",
	// 	SecretAccessKey: "rootroot",
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// _, err = s.Open("bucket1")
	// if err != nil {
	// 	panic(err)
	// }
}
