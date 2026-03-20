package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/rednafi/examples/cross-repository-transactions/bookstore"
	"github.com/rednafi/examples/cross-repository-transactions/sqlite"
)

func main() {
	db, err := sqlite.SetupDB("bookstore.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stores := bookstore.Stores{
		Books:  sqlite.NewBookStore(db),
		Orders: sqlite.NewOrderStore(db),
	}
	uow := sqlite.NewUoW(db)
	svc := bookstore.NewService(stores, uow)

	http.HandleFunc("POST /books", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
			Stock int    `json:"stock"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
			http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
			return
		}

		id, err := stores.Books.Create(r.Context(),
			bookstore.Book{Title: req.Title, Stock: req.Stock})
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bookstore.Book{ID: id, Title: req.Title, Stock: req.Stock})
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

	http.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BookID int64 `json:"book_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BookID == 0 {
			http.Error(w, `{"error":"book_id is required"}`, http.StatusBadRequest)
			return
		}

		order, err := svc.PlaceOrder(r.Context(), req.BookID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
