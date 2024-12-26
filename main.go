package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
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

	router := mux.NewRouter()

	router.HandleFunc("/", HomeHandler).Methods("GET")
	router.HandleFunc("/tasks", fetchTasks).Methods("GET")
	router.HandleFunc("/tasks", addTask).Methods("POST")
	router.HandleFunc("/getnewtaskform", getTaskForm).Methods("GET")
	router.HandleFunc("/gettaskupdateform/{id}", getTaskUpdateForm).Methods("GET")
	router.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

	if err := http.ListenAndServe(":3000", router); err != nil {
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

func getTaskUpdateForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := getTaskById(db, taskId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "updateTaskForm", task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskItem := r.FormValue("task")
	isDone := r.FormValue("done")

	var taskStatus bool

	switch strings.ToLower(isDone) {
	case "yes", "on":
		taskStatus = true
	case "no", "off":
		taskStatus = false
	default:
		// http.Error(w, "Invalid value for 'done'", http.StatusBadRequest)
		// return
		taskStatus = false
	}

	task := Task{
		Id:   taskId,
		Task: taskItem,
		Done: taskStatus,
	}

	query := "UPDATE tasks SET task= ?, done = ? WHERE id = ?"

	result, err := db.Exec(query, task.Task, task.Done, task.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No rows updated", http.StatusNotFound)
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

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM tasks WHERE id = ?"

	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskId)
	if err != nil {
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

func getTaskById(db *sql.DB, id int) (*Task, error) {
	query := "SELECT id, task, done FROM tasks WHERE id = ?"

	var task Task

	row := db.QueryRow(query, id)
	err := row.Scan(&task.Id, &task.Task, &task.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with id %d not found", id)
		}
		return nil, err
	}

	return &task, nil
}
