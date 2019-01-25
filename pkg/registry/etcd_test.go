package registry

import (
	"testing"
	"time"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
	endpoints      = []string{"localhost:2379"}
)

func TestEtcdRegister(t *testing.T) {
	es := NewEtcdRegistry(dialTimeout, requestTimeout, endpoints)
	myapp := NewService("task", "1.1.0", "172.16.1.12", 8889)

	for i := 0; i < 1; i++ {
		es.Register(myapp)
	}

	for i := 0; i < 1; i++ {
		if ss, err := es.GetService("task1.1.0", ""); err != nil {
			t.Error(err)
		} else {
			t.Logf("yyyyy: %#v\n", ss)
			for _, s := range ss {
				t.Logf("yyyy service: %#+v\n", s)
			}
		}
	}

	//es.Deregister(myapp)

	if ss, err := es.ListServices(); err != nil {
		t.Error(err)
	} else {
		t.Logf("all services: %#+v\n", ss)
		for _, s := range ss {
			t.Logf("service: %#+v\n", s)
		}
	}
	es.Close()
}
