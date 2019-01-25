package registry

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
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

func (c *EtcdRegistry) RegisterWithTTL(s *Service, timeTTL time.Duration) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}

	// use first node
	node := s.Nodes[0]
	val := strings.Join(s.GetTags(), ",")
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()

	// ttl-second
	resp, _ := c.client.Grant(ctx, timeTTL.Nanoseconds())

	// get and put
	_, err := c.client.Get(ctx, node.Id)
	if err != nil {
		if err == rpctypes.ErrKeyNotFound {
			if _, err := c.client.Put(ctx, node.Id, val, clientv3.WithLease(resp.ID)); err != nil {
				log.Errorf("set service %s etcd3 failed: %v", node.Id, err)
				return err
			}
		} else {
			log.Infof("get failed: Id:%s, error: %v", node.Id, err)
			return err
		}
	} else {
		// refresh
		if _, err := c.client.Put(ctx, node.Id, val, clientv3.WithLease(resp.ID)); err != nil {
			log.Errorf("refresh service %s failed: %v", node.Id, err)
			return err
		}
	}
	return nil
}

func (c *EtcdRegistry) Register(s *Service) error {
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

	key := fmt.Sprintf("/%s/%s", DefaultPrefixService, name)
	resp, err := c.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil || resp == nil {
		log.Errorf("etcd GetWithPrefix error:%v", err)
		return nil, err
	}

	serviceMap := map[string]*Service{}
	//resList := make([]*Service, len(resp.Kvs))
	for i, ev := range resp.Kvs {
		log.Infof("ev %d: %#v \n", i, ev)

		tags := strings.Split(string(ev.Value), ",")
		keys := strings.Split(string(ev.Key), "/")

		fmt.Println("keys :", keys)
		fmt.Println("tags :", tags)
		if len(keys) < 4 || keys[2] != name {
			continue
		}

		version := tagsVersion(tags)
		address := keys[3]
		// use node address
		if len(address) == 0 {
			address = tagsHost(tags)
		}

		key := tagsCheck(tags, tag)
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

		port, _ := strconv.Atoi(keys[4])
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

	key := fmt.Sprintf("/%s", DefaultPrefixService)
	resp, err := c.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil || resp == nil {
		log.Errorf("etcd GetWithPrefix error:%v", err)
		return nil, err
	}

	services := make([]*Service, 0, len(resp.Kvs))
	for _, ev := range resp.Kvs {
		keys := strings.Split(string(ev.Key), "/")
		fmt.Println("keysdsddd: ", keys)
		if len(keys) > 3 {
			services = append(services, &Service{Name: keys[2]})
		}
	}
	return services, nil
}

func (c *EtcdRegistry) Close() error {
	if c != nil {
		c.client.Close()
	}
	return nil
}
