package main

import (
	"hotreload/internal/watcher"
	"log"
)

func main() {
	// ... flag parsing ...
	w, err := watcher.New(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	go func() {
		for event := range w.Events() {
			log.Println("event:", event)
		}
	}()
	go func() {
		for err := range w.Errors() {
			log.Println("watcher error:", err)
		}
	}()

	// Keep running
	select {}
}
