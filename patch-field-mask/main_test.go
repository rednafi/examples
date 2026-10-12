package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPatchUser(t *testing.T) {
	tests := []struct {
		name string
		mask string
		body string
		want User
	}{
		{
			name: "keep address",
			mask: "name",
			body: `{"name": "Bob"}`,
			want: User{ID: "1", Name: "Bob", Address: "Old Street 1"},
		},
		{
			name: "clear address",
			mask: "address",
			body: `{}`,
			want: User{ID: "1", Name: "Alice", Address: ""},
		},
		{
			name: "replace address",
			mask: "address",
			body: `{"address": "New Street 2"}`,
			want: User{ID: "1", Name: "Alice", Address: "New Street 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(User{ID: "1", Name: "Alice", Address: "Old Street 1"})
			srv := httptest.NewServer(s.Handler())
			defer srv.Close()

			req, err := http.NewRequest(
				http.MethodPatch,
				srv.URL+"/users/1?update_mask="+tt.mask,
				strings.NewReader(tt.body),
			)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			var got User
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
