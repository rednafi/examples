// Command patch-field-mask serves a PATCH endpoint that uses an update mask to
// decide which fields to change.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Server struct {
	mu    sync.Mutex
	users map[string]User
}

func NewServer(users ...User) *Server {
	s := &Server{users: make(map[string]User)}
	for _, u := range users {
		s.users[u.ID] = u
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /users/{id}", s.patchUser)
	return mux
}

func (s *Server) patchUser(w http.ResponseWriter, r *http.Request) {
	mask := r.URL.Query().Get("update_mask")
	if mask == "" {
		http.Error(w, "update_mask is required", http.StatusBadRequest)
		return
	}

	var in User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[r.PathValue("id")]
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	for path := range strings.SplitSeq(mask, ",") {
		switch path {
		case "name":
			u.Name = in.Name
		case "address":
			u.Address = in.Address
		default:
			http.Error(w, "unknown field in update_mask: "+path,
				http.StatusBadRequest)
			return
		}
	}
	s.users[u.ID] = u

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func main() {
	s := NewServer(User{ID: "1", Name: "Alice", Address: "Old Street 1"})
	log.Println("listening on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", s.Handler()))
}
