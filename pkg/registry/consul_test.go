// Copyright 2018 The huayulei_2003@hotmail.com Authors
// This file is part of the airfk library.
//
// The airfk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The airfk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the airfk library. If not, see <http://www.gnu.org/licenses/>.
package registry

import (
	"testing"
)

var (
	hosts = "localhost:8500"
)

func TestConsulRegister(t *testing.T) {
	es, _ := NewConsulRegistry(hosts, 0)
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
				t.Logf("yyyy nodes: %#+v\n", s.Nodes[0])
			}
		}
	}

	es.Deregister(myapp)

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
