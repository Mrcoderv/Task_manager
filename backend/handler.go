package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func getTasks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {

	var task Task

	json.NewDecoder(r.Body).Decode(&task)

	task.ID = len(tasks) + 1

	tasks = append(tasks, task)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(task)
}

// DELETE /tasks/:id
func deleteTask(w http.ResponseWriter, r *http.Request) {

	idText := strings.TrimPrefix(r.URL.Path, "/tasks/")

	id, err := strconv.Atoi(idText)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for i, task := range tasks {

		if task.ID == id {

			tasks = append(tasks[:i], tasks[i+1:]...)

			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
