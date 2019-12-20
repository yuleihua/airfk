package zoo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	git "gopkg.in/src-d/go-git.v4"
)

type Project struct {
	Name       string
	Website    string
	ProjectDir string
	RelDir     string
	LibDir     string
	IsRemove   bool
	Folder     ZooFolder
}

var nodeFileList = map[string][]string{
	"apis":    {"@@@@@.proto"},
	"config":  {"conf.go"},
	"metrics": {"metrics.go"},
	"version": {"version.go"},
}

var distFileList = map[string][]string{
	"etc": {"@@@@@.ini"},
	"bin": {""},
}

func NewProject(website, name, lib string, isRemove bool) *Project {
	rPath := ProjectPath(website, name)
	return &Project{
		Name:       name,
		Website:    website,
		ProjectDir: rPath,
		LibDir:     lib,
		IsRemove:   isRemove,
		RelDir:     fmt.Sprintf("%s/%s", website, name),
	}
}

func (p *Project) AddFile(templatePath string) error {
	err := os.MkdirAll(p.ProjectDir, os.ModePerm)
	if err != nil {
		return err
	}
	return p.Folder.render(templatePath, p)
}

func (p *Project) Write(templatePath string) error {
	if p.IsRemove {
		os.RemoveAll(p.ProjectDir)
	}

	err := os.MkdirAll(p.ProjectDir, os.ModePerm)
	if err != nil {
		return err
	}
	return p.Folder.render(templatePath, p)
}

func (p *Project) Clone(gitList []string) error {
	for _, url := range gitList {
		if url == "" {
			continue
		}

		dst := filepath.Join(srcPath(), url)
		if fileExist(dst) {
			continue
		}
		//f := filepath.Join(p.ProjectDir, k)
		//if err := gitClone(v, f); err != nil {
		//	return err
		//}
		c := fmt.Sprintf("go get %s", url)
		_, err := exec.Command(c).Output()
		if err != nil {
			fmt.Println("go get  error:", err)
			return err
		}
	}
	return nil
}

func (p Project) Test(url string) error {
	return nil
}

func (p Project) Build() error {
	c := fmt.Sprintf("cd %s && make build", p.ProjectDir)
	out, err := exec.Command(c).Output()
	if err != nil {
		fmt.Println("build error:", err)
		return err
	}
	fmt.Println("build out:\n", out)
	return nil
}

// CamelCaseName
func (p Project) CamelCaseName() string {
	return strcase.ToCamel(p.Name)
}

// SnakeCaseName
func (p Project) SnakeCaseName() string {
	return strings.Replace(strcase.ToSnake(p.Name), "-", "_", -1)
}

// DNSName
func (p Project) DNSName() string {
	return strings.Replace(strcase.ToSnake(p.Name), "_", "-", -1)
}

// LibName pkg url
func (p Project) LibName() string {
	return p.LibDir
}

func ProjectPath(website, name string) string {
	if website == "" {
		x, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			return ""
		}
		return filepath.Join(x, name)
	}

	srcPath := srcPath()
	return filepath.Join(srcPath, website, name)
}

func srcPath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src")
}

func gitClone(url, store string) error {
	_, err := git.PlainClone(store, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("git clone error:", err)
		return err
	}
	return nil
}

func fileExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
