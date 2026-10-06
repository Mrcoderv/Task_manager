package main


// this temporary memory storage will be replaced with a database in the future
// Task represents a task in the task manager..


 
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
