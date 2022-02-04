// Package form provides ...
package form

type errors map[string][]string

func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

func (e errors) Get(field string) string {
	em := e[field]
	if len(em) == 0 {
		return ""
	}
	return em[0]
}
