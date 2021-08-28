package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	filename    *string = flag.String("in", "", "name of the file you want to run the ast-replacer on")
	outFilename *string = flag.String("out", "", "name of the file you want the result to be written to")
)

func init() {
	flag.Parse()

	if *filename == "" || *outFilename == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func logic() error {
	fileContent, err := ioutil.ReadFile(*filename)
	if err != nil {
		return err
	}

	newFileContent, err := replace(*filename, string(fileContent))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(*outFilename, newFileContent, 0700)
	if err != nil {
		return err
	}

	newFileContent, err = exec.Command("goimports", *outFilename).Output()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(*outFilename, newFileContent, 0700) // TODO: should not need to write the file two times
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
