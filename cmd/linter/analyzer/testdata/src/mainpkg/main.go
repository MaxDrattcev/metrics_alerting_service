package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed in main.main")
	os.Exit(0)
}

func helper() {
	log.Fatal("forbidden") // want "log.Fatal is forbidden outside main.main"
	os.Exit(2)             // want "os.Exit is forbidden outside main.main"
}
