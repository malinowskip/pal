package main

import (
	"log"
	"os"
	"github.com/malinowskip/pal/app"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
