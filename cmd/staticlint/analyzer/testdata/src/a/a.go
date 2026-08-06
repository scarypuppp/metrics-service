package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("ok") // здесь разрешено
	os.Exit(1)      // здесь разрешено
}

func other() {
	panic("boom")   // want "usage of panic is forbidden"
	log.Fatal("no") // want "calling log.Fatal is forbidden outside of main.main"
	os.Exit(1)      // want "calling os.Exit is forbidden outside of main.main"
}
