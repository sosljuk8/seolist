package main

import (
	"io"
	"log"
	"os"
)

// ReadHTMLFromFile reads HTML from file and returns full text as a string
func ReadHTMLFromFile(filename string) (string, error) {

	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Error closing file", err.Error())
		}
	}(f)

	b, err := io.ReadAll(f)

	return string(b), err
}
