package main

import (
	"log"
	"os"
	"pal/app"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
