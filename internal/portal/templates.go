package portal

import (
	"html/template"
)

var rootTemplate *template.Template

func ImportTemplates(dir string) error {
	var err error

	rootTemplate, err = template.ParseFiles(
		dir+"/students.html",
		dir+"/student.html")

	if err != nil {
		return err
	}

	return nil
}
