package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("error in main")

	os.Exit(1)
}

func badFunction() {
	panic("something went wrong") // want "panic usage is forbidden"

	log.Fatal("error outside main") // want "log.Fatal must be called only in main.main"

	os.Exit(1) // want "os.Exit must be called only in main.main"
}

func anotherFunction() {
	func() {
		panic("panic in anonymous") // want "panic usage is forbidden"

		log.Fatal("fatal in anonymous") // want "log.Fatal must be called only in main.main"
	}()
}

func init() {
	log.Fatal("fatal in init") // want "log.Fatal must be called only in main.main"
}
