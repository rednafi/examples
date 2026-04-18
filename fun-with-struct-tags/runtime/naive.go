// Package runtime is the runtime-reflection side of the blog post.
// Naive and registry-based validators live here.
package runtime

import (
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
)

func ValidateNaive(s any) error {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := range t.NumField() {
		f, name := v.Field(i), t.Field(i).Name
		tag := t.Field(i).Tag.Get("check")

		for rule := range strings.SplitSeq(tag, ",") {
			head, arg, _ := strings.Cut(rule, "=")

			switch head {
			case "required":
				if f.IsZero() {
					return fmt.Errorf("%s: required", name)
				}
			case "min":
				n, _ := strconv.Atoi(arg)
				if len(f.String()) < n {
					return fmt.Errorf("%s: min %d", name, n)
				}
			case "email":
				if _, err := mail.ParseAddress(f.String()); err != nil {
					return fmt.Errorf("%s: %w", name, err)
				}
			}
		}
	}
	return nil
}
