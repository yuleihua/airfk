package zoo

import (
	"fmt"
	"testing"
)

func TestProject(t *testing.T) {
	webName := "airman.com"
	name := "sanzoTest"
	rPath := ProjectPath(webName, name)
	p := &Project{
		Name:       name,
		Website:    webName,
		ProjectDir: rPath,
		RelDir:     fmt.Sprintf("%s/%s", webName, name),
	}

	t.Logf("CamelCaseName:%s\n", p.CamelCaseName())
	t.Logf("SnakeCaseName:%s\n", p.SnakeCaseName())
	t.Logf("DNSName:%s\n", p.DNSName())
	t.Logf("ProjectDir:%s\n", rPath)
	t.Logf("RelDir:%s\n", p.RelDir)
}

func TestGitClone(t *testing.T) {
	url := "https://github.com/jnewmano/grpc-json-proxy"
	webName := "airman.com"
	name := "sanzoTest"

	rPath := ProjectPath(webName, name)
	if err := gitClone(url, rPath); err != nil {
		t.Fatal(err)
	}
}
