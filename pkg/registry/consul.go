package registry

import (
	"errors"
	"fmt"
	"net"
	"time"

	consul "github.com/hashicorp/consul/api"
)

const (
	DefaultInterval   = 10 * time.Second
	DefaultTagService = "service"
)

type ConsulRegistry struct {
	addr         string
	client       *consul.Client
	timeout      time.Duration
	queryOptions *consul.QueryOptions
}

func NewConsulRegistry(address string, timeout time.Duration) (*ConsulRegistry, error) {
	r := &ConsulRegistry{
		queryOptions: &consul.QueryOptions{
			AllowStale: true,
		},
	}
	if err := configure(r, address, timeout); err != nil {
		return nil, err
	}
	return r, nil
}

func configure(c *ConsulRegistry, address string, timeout time.Duration) error {
	// use default config
	config := consul.DefaultConfig()

	// check if there are any address
	if address != "" {
		addr, port, err := net.SplitHostPort(address)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "8500"
			addr = address
			config.Address = fmt.Sprintf("%s:%s", addr, port)
		} else if err == nil {
			config.Address = fmt.Sprintf("%s:%s", addr, port)
		}
	}

	// set timeout
	//if timeout > 0 {
	//	config.HttpClient.Timeout = timeout
	//}

	// create the client
	client, err := consul.NewClient(config)
	if err != nil {
		return err
	}

	// set address/client
	c.addr = config.Address
	c.client = client
	c.timeout = timeout
	return nil
}

func getDeregisterTTL(t time.Duration) time.Duration {
	deregTTL := t + DefaultInterval

	// consul has a minimum timeout on deregistration of 1 minute.
	if t < 2*time.Minute {
		deregTTL = 2 * time.Minute
	}
	return deregTTL
}

func (c *ConsulRegistry) RegisterWithTTL(s *Service, timeTTL time.Duration) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}

	// use first node
	node := s.Nodes[0]

	// full re-register
	if err := c.client.Agent().PassTTL("service:"+node.Id, ""); err == nil {
		return nil
	}

	deregTTL := getDeregisterTTL(timeTTL)
	check := &consul.AgentServiceCheck{
		Name:                           s.Name,
		Notes:                          node.Id,
		TTL:                            fmt.Sprintf("%v", timeTTL),
		DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
	}

	// register the service
	asr := &consul.AgentServiceRegistration{
		ID:      node.Id,
		Name:    s.Name,
		Tags:    s.GetTags(),
		Port:    node.Port,
		Address: node.Host,
		Check:   check,
	}

	//// Specify consul connect
	//if c.connect {
	//	asr.Connect = &consul.AgentServiceConnect{
	//		Native: true,
	//	}
	//}

	if err := c.client.Agent().ServiceRegister(asr); err != nil {
		return err
	}

	// if the TTL is 0 we don't mess with the checks
	if timeTTL == time.Duration(0) {
		return nil
	}
	return c.client.Agent().PassTTL("service:"+node.Id, "")
}

// default is 30s.
func (c *ConsulRegistry) Register(s *Service) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}

	// use first node
	node := s.Nodes[0]

	// full re-register
	if err := c.client.Agent().PassTTL("service:"+node.Id, ""); err == nil {
		return nil
	}

	unregTTL := getDeregisterTTL(0)
	healthCheck := fmt.Sprintf("http://%v:%v/%v", node.Host, node.Port, s.Check)
	if s.Check == "" {
		healthCheck = fmt.Sprintf("http://%v:%v", node.Host, node.Port)
	}
	check := &consul.AgentServiceCheck{
		Name:                           s.Name,
		Notes:                          node.Id,
		Interval:                       "10s",
		HTTP:                           healthCheck,
		Timeout:                        "30s",
		DeregisterCriticalServiceAfter: fmt.Sprintf("%v", unregTTL),
	}

	// register the service
	asr := &consul.AgentServiceRegistration{
		ID:      node.Id,
		Name:    s.Name,
		Tags:    s.GetTags(),
		Port:    node.Port,
		Address: node.Host,
		Check:   check,
	}

	if err := c.client.Agent().ServiceRegister(asr); err != nil {
		return err
	}
	return nil
}

func (c *ConsulRegistry) Deregister(s *Service) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}
	node := s.Nodes[0]
	return c.client.Agent().ServiceDeregister(node.Id)
}

func (c *ConsulRegistry) GetService(name, tag string) ([]*Service, error) {
	rsp, _, err := c.client.Health().Service(name, tag, false, c.queryOptions)
	if err != nil {
		return nil, err
	}

	serviceMap := map[string]*Service{}
	for _, s := range rsp {
		if s.Service.Service != name {
			continue
		}

		var del bool
		for _, check := range s.Checks {
			// delete the node if the status is critical
			if check.Status == "critical" {
				del = true
				break
			}
		}

		// if delete then skip the node
		if del {
			continue
		}

		version := tagsVersion(s.Service.Tags)
		address := s.Service.Address
		// use node address
		if len(address) == 0 {
			address = s.Node.Address
		}

		key := tagsCheck(s.Service.Tags, tag)
		if tag == "" {
			key = version
		}
		svc, ok := serviceMap[key]
		if !ok {
			svc = &Service{
				Name:    s.Service.Service,
				Version: version,
				Tags:    s.Service.Tags,
			}
			serviceMap[key] = svc
		}

		svc.Nodes = append(svc.Nodes, &Node{
			Id:   s.Service.ID,
			Host: address,
			Port: s.Service.Port,
		})
	}

	var services []*Service
	for _, service := range serviceMap {
		services = append(services, service)
	}
	return services, nil
}

func (c *ConsulRegistry) ListServices() ([]*Service, error) {
	rsp, _, err := c.client.Catalog().Services(c.queryOptions)
	if err != nil {
		return nil, err
	}

	var services []*Service

	for service := range rsp {
		services = append(services, &Service{Name: service})
	}
	return services, nil
}

func (c *ConsulRegistry) Close() error {
	return nil
}
