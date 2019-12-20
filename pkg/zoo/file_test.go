package zoo

import (
	"fmt"
	"os"
	"testing"
)

func TestFolder(t *testing.T) {
	webName := "airman.com"
	name := "sanzoTest"
	rPath := ProjectPath(webName, name)
	f := ZooFolder{Name: name, AbsPath: rPath}
	f.AddFile("main.go", "main.go.tpl")
	f.AddFile(".gitignore", "gitignore.tpl")

	p := &Project{
		Name:       name,
		Website:    webName,
		ProjectDir: rPath,
		RelDir:     fmt.Sprintf("%s/%s", webName, name),
		Folder:     f,
	}

	if err := os.MkdirAll(rPath, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := f.render("testdata", p); err != nil {
		t.Fatal()
	}
}
