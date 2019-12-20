package zoo

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	tpl "text/template"
)

type ZooFile struct {
	Name     string
	AbsPath  string
	Template string
}

type ZooFolder struct {
	Name    string
	AbsPath string

	files   []ZooFile
	folders []*ZooFolder
}

func (f *ZooFolder) AddFolder(name string) *ZooFolder {
	newF := &ZooFolder{
		Name:    name,
		AbsPath: filepath.Join(f.AbsPath, name),
	}
	f.folders = append(f.folders, newF)
	return newF
}

func (f *ZooFolder) AddFile(name, tmpl string) {
	f.files = append(f.files, ZooFile{
		Name:     name,
		Template: tmpl,
		AbsPath:  filepath.Join(f.AbsPath, name),
	})
}

func (f *ZooFolder) render(templatePath string, p *Project) error {
	for _, v := range f.files {
		t, err := tpl.ParseFiles(filepath.Join(templatePath, v.Template))
		if err != nil {
			return err
		}

		ZooFile, err := os.Create(v.AbsPath)
		if err != nil {
			return err
		}

		defer ZooFile.Close()

		if strings.Contains(v.AbsPath, ".go") {
			var out bytes.Buffer
			err = t.Execute(&out, p)
			if err != nil {
				log.Printf("Could not process template %s\n", v)
				return err
			}

			b, err := format.Source(out.Bytes())
			if err != nil {
				fmt.Print(string(out.Bytes()))
				log.Printf("\nCould not format Go ZooFile %s\n", v)
				return err
			}

			_, err = ZooFile.Write(b)
			if err != nil {
				return err
			}

		} else {
			err = t.Execute(ZooFile, p)
			if err != nil {
				return err
			}
		}
	}

	for _, v := range f.folders {
		err := os.Mkdir(v.AbsPath, os.ModePerm)
		if err != nil {
			return err
		}

		err = v.render(templatePath, p)
		if err != nil {
			return err
		}
	}
	return nil
}
