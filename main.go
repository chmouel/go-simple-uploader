package main

import (
	"log"

	"github.com/chmouel/go-simple-uploader/uploader"
)

func main() {
	err := uploader.Uploader()
	if err != nil {
		log.Fatal(err)
	}
}
