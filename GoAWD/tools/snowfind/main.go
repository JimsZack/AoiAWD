package main

import (
	"log"
	"goawd/tools/snowfind/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
