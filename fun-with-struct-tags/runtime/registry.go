package runtime

import (
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
)

type Rule func(f reflect.Value, arg string) error

var rules = map[string]Rule{
	"required": func(f reflect.Value, _ string) error {
		if f.IsZero() {
			return errors.New("required")
		}
		return nil
	},
	"min": func(f reflect.Value, arg string) error {
		n, _ := strconv.Atoi(arg)
		if len(f.String()) < n {
			return fmt.Errorf("min length %d", n)
		}
		return nil
	},
	"email": func(f reflect.Value, _ string) error {
		_, err := mail.ParseAddress(f.String())
		return err
	},
}

func Validate(s any) error {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := range t.NumField() {
		tag := t.Field(i).Tag.Get("check")
		for rule := range strings.SplitSeq(tag, ",") {
			head, arg, _ := strings.Cut(rule, "=")
			fn, ok := rules[head]
			if !ok {
				continue
			}
			if err := fn(v.Field(i), arg); err != nil {
				return fmt.Errorf("%s: %w", t.Field(i).Name, err)
			}
		}
	}
	return nil
}
