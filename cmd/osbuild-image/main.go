package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ondrejbudai/osbuild-image/internal/weldr-image"
)

type flags struct {
	imageType     string
	imagePath     string
	manifestPath  string
	logPath       string
	blueprintPath string
	keepArtifacts bool
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
	flag.StringVar(&flags.blueprintPath, "blueprint", "", "json or toml blueprint to be used (optional, if not specified, an empty blueprint will be used)")
	flag.StringVar(&flags.imageType, "type", "", "image type to be built")
	flag.StringVar(&flags.imagePath, "output", "", "path where the image will be saved")
	flag.StringVar(&flags.manifestPath, "output-manifest", "", "path where the manifest will be saved (optional, it's not saved if no path is given)")
	flag.StringVar(&flags.logPath, "output-log", "", "path where the log will be saved (optional, it's not saved if no path is given)")
	flag.BoolVar(&flags.keepArtifacts, "keep-artifacts", false, "whether osbuild-image should keep all the artifacts (blueprint and compose), false by default, that means osbuild-image cleans up after itself")
	flag.Parse()

	if err := validateFlags(&flags); err != nil {
		fmt.Fprintf(os.Stderr, "validation of arguments failed: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	imageFile, err := os.Create(flags.imagePath)
	if err != nil {
		log.Fatal("cannot open the output file: ", err)
	}

	var blueprint []byte
	if flags.blueprintPath != "" {
		blueprint, err = ioutil.ReadFile(flags.blueprintPath)
		if err != nil {
			log.Fatal("cannot read the blueprint file: ", err)
		}
	}

	req := &weldr_image.Request{
		Blueprint:     blueprint,
		ImageType:     flags.imageType,
		ImageWriter:   imageFile,
		ManifestPath:  flags.manifestPath,
		LogPath:       flags.logPath,
		KeepArtifacts: flags.keepArtifacts,
	}

	err = req.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "validation of the image request failed: %v\n", err)
		os.Exit(1)
	}

	err = req.Process()

	if err != nil {
		log.Fatal(err)
	}
}
