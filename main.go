package main

import (
	"encoding/json"
	"net/http"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
    "time"
    "github.com/gorilla/mux"
    "strings"
    "strconv"
)

type Todo struct {
    ID int `json:"id"`
    Title string `json:"title"`
    StartDate time.Time `json:"start_date"`
    EndDate *time.Time `json:"end_date"`
}

type ReqTodo struct {
    Title string `json:"title"`
    StartDate *time.Time `json:"start_date"`
    EndDate *time.Time `json:"end_date"`
}
    	
        
func main() {
    db, err := sql.Open(
    	"mysql",
    	"root:@tcp(localhost:3306)/gop?parseTime=true",
    )
    if (err != nil) {
        panic(err)
    }
    
    err = db.Ping()
    if err != nil {
    	panic(err)
    }
    
    r := mux.NewRouter()
	r.HandleFunc("/todos", getTodosHandler(db)).Methods("GET")
	r.HandleFunc("/todos", postTodoHandler(db)).Methods("POST")
	r.HandleFunc("/todos/{id}", putTodoHandler(db)).Methods("PUT")
    r.HandleFunc("/todos/{id}", deleteTodoHandler(db)).Methods("DELETE")
	http.ListenAndServe(":8080", r)
}

func getTodosHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        query := "select * from todos"
        
        rows, err := db.Query(query)
        if (err != nil) {
            panic(err)
        };
        defer rows.Close()
        
        var responses []Todo
        for rows.Next() {
            var t Todo
            
            err = rows.Scan(&t.ID, &t.Title, &t.StartDate, &t.EndDate)
        	if err != nil {
        		panic(err)
        	}
        	
        	responses = append(responses, t)
        
        }
    
        w.Header().Set("Content-Type", "application/json")
    	json.NewEncoder(w).Encode(responses)
    }
}

func postTodoHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    	contentType := r.Header.Get("Content-Type")

    	if !strings.HasPrefix(contentType, "application/json") {
    		http.Error(w, "Content-Type harus application/json", http.StatusUnsupportedMediaType)
    		return
    	}
    	
    	var reqTodo ReqTodo
    	
    	err := json.NewDecoder(r.Body).Decode(&reqTodo)
		if err != nil {
			http.Error(w, "body tidak valid", http.StatusBadRequest)
			return
		}
		
		var res sql.Result
		var err2 error
		if reqTodo.StartDate == nil {
		    query := "insert into todos (title, end_date) values (?, ?)"
            res, err2 = db.Exec(query, reqTodo.Title, reqTodo.EndDate)
		} else {
		    query := "insert into todos (title, start_date, end_date) values (?, ?, ?)"
            res, err2 = db.Exec(query, reqTodo.Title, reqTodo.StartDate, reqTodo.EndDate)
		}
    
        if (err2 != nil) {
            panic(err2)
        }
        
        id, err := res.LastInsertId()
        if err != nil {
        	panic(err)
        }
        
        row := db.QueryRow("select * from todos where id = ?", id)
        if (err != nil) {
            panic(err)
        };
        

        var t Todo
        
        err = row.Scan(&t.ID, &t.Title, &t.StartDate, &t.EndDate)
    	if err != nil {
    		panic(err)
    	}
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
    	json.NewEncoder(w).Encode(t)
    }
}

func putTodoHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        contentType := r.Header.Get("Content-Type")

    	if !strings.HasPrefix(contentType, "application/json") {
    		http.Error(w, "Content-Type harus application/json", http.StatusUnsupportedMediaType)
    		return
    	}
    	
        idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "id invalid", http.StatusBadRequest)
			return
		}
		
		var reqTodo ReqTodo
    	
    	err = json.NewDecoder(r.Body).Decode(&reqTodo)
		if err != nil {
			http.Error(w, "body tidak valid", http.StatusBadRequest)
			return
		}
		
		var sets []string
		var args []any
		
		if reqTodo.Title != "" {
        	sets = append(sets, "title = ?")
        	args = append(args, reqTodo.Title)
        }
        
        if reqTodo.StartDate != nil {
        	sets = append(sets, "start_date = ?")
        	args = append(args, reqTodo.StartDate)
        }
        
        if reqTodo.EndDate != nil {
        	sets = append(sets, "end_date = ?")
        	args = append(args, reqTodo.EndDate)
        }
        
        if len(sets) == 0 {
            http.Error(w, "tidak ada data untuk diupdate", http.StatusBadRequest)
			return
        }
        
        query := "update todos set " + strings.Join(sets, ", ") + " where id = ?"
        args = append(args, id)
    
        _, errres := db.Exec(query, args...)
        if errres != nil {
        	http.Error(w, "update gagal", 500)
        	return
        }
        
        row := db.QueryRow("select * from todos where id = ?", id)
        if (err != nil) {
            panic(err)
        };
        

        var t Todo
        
        err = row.Scan(&t.ID, &t.Title, &t.StartDate, &t.EndDate)
    	if err != nil {
    		panic(err)
    	}
        
        w.Header().Set("Content-Type", "application/json")
    	json.NewEncoder(w).Encode(t)
    }
}
func deleteTodoHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    	idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "id invalid", http.StatusBadRequest)
			return
		}
		
		query := "delete from todos where id = ?"
		
		_, err2 := db.Exec(query, id)
        if (err2 != nil) {
            panic(err2)
        }
        
        w.WriteHeader(http.StatusNoContent)
    }
}