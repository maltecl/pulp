package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	inFilename  *string = flag.String("in", "", "name of the file you want to run the ast-replacer on")
	outFilename *string = flag.String("out", "", "name of the file you want the result to be written to")
	verbose     *bool   = flag.Bool("v", false, "if this value is true, the program is more chatty")
)

func init() {
	flag.Parse()

	if *inFilename == "" || *outFilename == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func logic() error {
	fileContent, err := ioutil.ReadFile(*inFilename)
	if err != nil {
		return err
	}

	newFileContent, err := replace(*inFilename, string(fileContent))
	if err != nil {
		return err
	}

	if len(newFileContent) == 0 {
		return fmt.Errorf("something went wrong. This probably means, that you have an error in %s", *inFilename)
	}

	err = ioutil.WriteFile(*outFilename, newFileContent, 0700)
	if err != nil {
		return err
	}

	withImports, err := exec.Command("goimports", *outFilename).Output() // ignore the error, so that when goimports is not installed it's okay
	if err == nil {
		newFileContent = withImports
	}

	return ioutil.WriteFile(*outFilename, newFileContent, 0700) // TODO: should not need to write the file two times
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
