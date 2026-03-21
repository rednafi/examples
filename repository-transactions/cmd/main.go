package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/rednafi/examples/repository-transactions/book"
	"github.com/rednafi/examples/repository-transactions/sqlite"
)

func main() {
	db, err := sqlite.SetupDB("books.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := sqlite.NewStore(db)
	svc := book.NewService(store)

	http.HandleFunc("POST /books", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
			http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
			return
		}

		book, err := svc.RegisterBook(r.Context(), req.Title)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})

	http.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}

		book, err := svc.GetBook(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
