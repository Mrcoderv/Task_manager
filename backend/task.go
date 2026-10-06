package main

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var tasks = []Task{
	{
		ID:        1,
		Title:     "Learning Js",
		Completed: false,
	},
	{
		ID:        2,
		Title:     "Learning  go",
		Completed: false,
	},
}
