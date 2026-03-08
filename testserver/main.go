package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"hotreload/testserver/helpers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("testserver/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			PID      int
			Time     string
			ExtraMsg string
		}{
			PID:      os.Getpid(),
			Time:     time.Now().Format(time.RFC1123),
			ExtraMsg: helpers.GetMessage(),
		}

		t.Execute(w, data)
	})
	
	log.Printf("Server starting on :%s (pid %d) - HOTRELOAD WORKS!", port, os.Getpid())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
