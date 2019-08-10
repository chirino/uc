package utils

import (
	"bytes"
	"text/template"
)

func ApplyTemplate(tmpl string, scope interface{}) ([]byte, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	err = t.Execute(buffer, scope)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
