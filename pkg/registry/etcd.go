package registry

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNilKey   = errors.New("key is null")
	ErrNilValue = errors.New("value is null")
)

// etcd clientv3
//
type EtcdRegistry struct {
	dailTimeOut time.Duration
	reqTimeout  time.Duration
	endPoints   []string
	client      *clientv3.Client
}

func NewEtcdRegistry(tmDail, tmReq time.Duration, endPoints []string) *EtcdRegistry {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endPoints,
		DialTimeout: tmDail,
	})

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &EtcdRegistry{
		dailTimeOut: tmDail,
		reqTimeout:  tmReq,
		endPoints:   endPoints,
		client:      cli,
	}
}

func (c *EtcdRegistry) Register(s *Service, timeTTL time.Duration) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}

	// use first node
	node := s.Nodes[0]
	val := strings.Join(s.GetTags(), ",")
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()

	if _, err := c.client.Put(ctx, node.Id, val); err != nil {
		log.Errorf("etcd put error:%v", err)
		return err
	}
	return nil
}

func (c *EtcdRegistry) Deregister(s *Service) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}
	node := s.Nodes[0]
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()

	if _, err := c.client.Delete(ctx, node.Id); err != nil {
		log.Errorf("etcd Delete error:%v", err)
		return err
	}
	return nil
}

func (c *EtcdRegistry) GetService(name, tag string) ([]*Service, error) {
	if name == "" {
		return nil, ErrNilKey
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()

	resp, err := c.client.Get(ctx, name, clientv3.WithPrefix())
	if err != nil || resp == nil {
		log.Errorf("etcd GetWithPrefix error:%v", err)
		return nil, err
	}

	serviceMap := map[string]*Service{}
	//resList := make([]*Service, len(resp.Kvs))
	for i, ev := range resp.Kvs {
		log.Infof("ev %d: %#v \n", i, ev)

		tags := strings.Split(string(ev.Value), ",")
		keys := strings.Split(string(ev.Key), ":")
		if keys[0] != name {
			continue
		}

		version := tagsVersion(tags)
		address := keys[1]
		// use node address
		if len(address) == 0 {
			address = tagsHost(tags)
		}

		key := tag
		if tag == "" {
			key = version
		}
		svc, ok := serviceMap[key]
		if !ok {
			svc = &Service{
				Name:    name,
				Version: version,
				Tags:    tags,
			}
			serviceMap[key] = svc
		}

		var port int
		if len(keys) > 2 {
			port, _ = strconv.Atoi(keys[2])
		}
		svc.Nodes = append(svc.Nodes, &Node{
			Id:   string(ev.Key),
			Host: address,
			Port: port,
		})
	}

	var services []*Service
	for _, service := range serviceMap {
		services = append(services, service)
	}
	return services, nil
}

func (c *EtcdRegistry) ListServices() ([]*Service, error) {

	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()

	resp, err := c.client.Get(ctx, "/", clientv3.WithPrefix())
	if err != nil || resp == nil {
		log.Errorf("etcd GetWithPrefix error:%v", err)
		return nil, err
	}

	services := make([]*Service, 0, len(resp.Kvs))
	for _, ev := range resp.Kvs {
		keys := strings.Split(string(ev.Key), ":")
		services = append(services, &Service{Name: keys[0], Version: keys[1]})
	}
	return services, nil
}

func (c *EtcdRegistry) Close() error {
	if c != nil {
		c.client.Close()
	}
	return nil
}
