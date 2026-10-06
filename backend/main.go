package main

import (
	"fmt"
	"net/http"
)

func main() {

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			getTasks(w, r)
			return
		}

		if r.Method == http.MethodPost {
			createTask(w, r)
			return
		}

		http.Error(w, " not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodDelete {
			deleteTask(w, r)
			return
		}

		http.Error(w, " not allowed", http.StatusMethodNotAllowed)
	})

	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	fmt.Println("Server running on http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
