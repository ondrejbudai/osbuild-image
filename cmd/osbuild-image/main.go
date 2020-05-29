package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ondrejbudai/osbuild-image/internal/weldr-image"
)

type flags struct {
	imageType string
	imagePath string
}

func validateFlags(flags *flags) error {
	// do not validate imageType, weldr_image can return all the possible values
	if flags.imagePath == "" {
		return errors.New("image path cannot be empty")
	}
	return nil
}

func main() {
	var flags flags
	flag.StringVar(&flags.imageType, "type", "", "image type to be built")
	flag.StringVar(&flags.imagePath, "output", "", "filename of the produced image")
	flag.Parse()

	if err := validateFlags(&flags); err != nil {
		fmt.Fprintf(os.Stderr, "validation of arguments failed: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	imageFile, err := os.Create(flags.imagePath)

	if err != nil {
		log.Fatal("cannot open the output file", err)
	}

	req := &weldr_image.Request{
		ImageType:   flags.imageType,
		ImageWriter: imageFile,
	}

	err = req.Validate()
	if err != nil {
		log.Fatal("validation of the inputs failed: ", err)
	}

	err = req.Process()

	if err != nil {
		log.Fatal(err)
	}
}
