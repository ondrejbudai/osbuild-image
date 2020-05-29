package main

import (
	"flag"
	"log"
	"os"

	"github.com/ondrejbudai/osbuild-image/internal/weldr-image"
)

func main() {
	imageType := flag.String("type", "", "image type to be built")
	output := flag.String("output", "", "filename of the produced image")
	flag.Parse()

	imageFile, err := os.Create(*output)

	if err != nil {
		log.Fatal("cannot open the output file", err)
	}

	req := &weldr_image.Request{
		ImageType:   *imageType,
		ImageWriter: imageFile,
	}

	err = req.Process()

	if err != nil {
		log.Fatal(err)
	}
}
