// Package ejdemo is the easyjson side of the blog post.
// Running `go generate ./...` here produces user_easyjson.go, which
// contains the concrete MarshalJSON/UnmarshalJSON methods shown in the
// build-time section.
package ejdemo

//go:generate go run github.com/mailru/easyjson/easyjson -all user.go

//easyjson:json
type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Admin bool   `json:"-"`
}
