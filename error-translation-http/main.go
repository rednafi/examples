package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rednafi/examples/error-translation-http/sqlite"
	"github.com/rednafi/examples/error-translation-http/user"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	store := sqlite.NewUserStore(db)
	svc := user.NewService(store)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		u, err := svc.GetUser(r.Context(), id)
		if err != nil {
			slog.ErrorContext(r.Context(), "get user failed",
				"user_id", id,
				"err", err,
			)
			writeError(w, err)
			return
		}
		json.NewEncoder(w).Encode(u)
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		u, err := svc.CreateUser(r.Context(), req.Name, req.Email)
		if err != nil {
			slog.ErrorContext(r.Context(), "create user failed",
				"name", req.Name,
				"email", req.Email,
				"err", err,
			)
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	})

	fmt.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, user.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, user.ErrConflict):
		http.Error(w, "conflict", http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
