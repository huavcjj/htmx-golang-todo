package main

import (
	"database/sql"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var tmpl *template.Template

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))

}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:root@(127.0.0.1:3333)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}
func main() {

	initDB()
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("GET /tasks", fetchTasks)
	mux.HandleFunc("POST /tasks", addTask)
	mux.HandleFunc("/getnewtaskform", getTaskForm)

	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}

// Routes
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fetchTasks(w http.ResponseWriter, r *http.Request) {

	todos, _ := getTasks(db)

	if err := tmpl.ExecuteTemplate(w, "todoList", todos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func getTaskForm(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "addTaskForm", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addTask(w http.ResponseWriter, r *http.Request) {
	task := r.FormValue("task")
	if task == "" {
		http.Error(w, "Task is required", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO tasks (task) VALUES (?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	todos, err := getTasks(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "todoList", todos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Utility
func getTasks(db *sql.DB) ([]Task, error) {
	query := "SELECT id, task, done FROM tasks"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var todo Task
		if err := rows.Scan(&todo.Id, &todo.Task, &todo.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}
