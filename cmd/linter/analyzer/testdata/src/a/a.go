package a

import (
	"log"
	"os"
)

func bad() {
	panic("boom")  // want "avoid using built-in panic"
	log.Fatal("x") // want "log.Fatal is forbidden outside main.main"
	os.Exit(1)     // want "os.Exit is forbidden outside main.main"
}

func shadowedPanicIsOk() {
	panic := func(string) {}
	panic("not builtin")
}
