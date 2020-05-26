package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

type TodoModel struct {
	Id        int `gorm:"primary_key"`
	Details   string
	Completed bool
}

var db, err = gorm.Open("sqlite3", "test4.db")

func CreateItem(w http.ResponseWriter, r *http.Request) {
	details := r.FormValue("details")
	todo := TodoModel{Details: details, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Value)

}
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := GetItem(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		todo := &TodoModel{}
		db.First(todo, id)
		todo.Completed = completed
		db.Save(todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	err := GetItem(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "Record Not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
		todo := &TodoModel{}
		db.First(todo, id)
		db.Delete(todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetItem(Id int) bool {
	todo := &TodoModel{}
	result := db.First(todo, Id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}
	return true
}

func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get completed TodoItems")
	completedTodoItems := GetTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func GetIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Incomplete TodoItems")
	IncompleteTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IncompleteTodoItems)
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoModel
	fmt.Printf("%v", todos)
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func main() {
	db.AutoMigrate(&TodoModel{})

	defer db.Close()

	log.Info("Starting server")
	r := mux.NewRouter()
	r.HandleFunc("/todo-completed", GetCompletedItems).Methods("GET")
	r.HandleFunc("/todo-incomplete", GetIncompleteItems).Methods("GET")
	r.HandleFunc("/todo", CreateItem).Methods("POST")
	r.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	r.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")
	http.ListenAndServe(":8080", r)
}
