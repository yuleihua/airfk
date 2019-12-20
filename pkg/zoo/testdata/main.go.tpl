package main

import (
	"fmt"
)

type test{{ .CamelCaseName }}Server struct {
     Id string
}

func main() {
	s := &test{{ .CamelCaseName }}Server{
	    Id:"{{ .SnakeCaseName }}Register{{ .CamelCaseName }}",
	}

	fmt.Println("{{ .Name }}")
	fmt.Println("test 1:", s.Id)

	s.Id = "{{ .RelDir }}/{{ .Name }}/pkg/log"
	fmt.Println("test 2:", s.Id)
}
