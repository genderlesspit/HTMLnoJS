package main

import (
	"flag"
	"log"

	"htmlnojs/setup"
)

func main() {
	directory := flag.String("directory", ".", "Project directory to setup")
	flag.Parse()

	err := setup.Run(*directory)
	if err != nil {
		log.Fatal(err)
	}
}