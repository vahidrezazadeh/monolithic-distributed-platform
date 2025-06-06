package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

type NodeInfo struct {
	ID       string `json:"id"`
	HTTPAddr string `json:"http_addr"`
	GRPCAddr string `json:"grpc_addr"`
	IsLeader bool   `json:"is_leader"`
	LastSeen int64  `json:"last_seen"`
	Version  string `json:"version"`
}

type Cluster struct {
	mlist      *memberlist.Memberlist
	localNode  *NodeInfo
	nodes      map[string]*NodeInfo
	nodesMutex sync.RWMutex
	config     Config
	stopChan   chan struct{}
}

func NewCluster(cfg Config) (*Cluster, error) {
	mlConfig := memberlist.DefaultLANConfig()
	mlConfig.Name = cfg.InstanceID
	mlConfig.BindPort = cfg.ClusterPort
	mlConfig.AdvertisePort = cfg.ClusterPort
	mlConfig.LogOutput = nil 

	mlist, err := memberlist.Create(mlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create memberlist: %v", err)
	}

	if len(cfg.JoinAddresses) > 0 {
		_, err = mlist.Join(cfg.JoinAddresses)
		if err != nil {
			log.Printf("Failed to join cluster: %v", err)
		} else {
			log.Printf("Joined cluster via %v", cfg.JoinAddresses)
		}
	}

	cluster := &Cluster{
		mlist:    mlist,
		config:   cfg,
		nodes:    make(map[string]*NodeInfo),
		stopChan: make(chan struct{}),
		localNode: &NodeInfo{
			ID:       cfg.InstanceID,
			HTTPAddr: fmt.Sprintf("http://%s:%d", getLocalIP(), cfg.HTTPPort),
			GRPCAddr: fmt.Sprintf("%s:%d", getLocalIP(), cfg.GRPCPort),
			Version:  version,
		},
	}

	go cluster.broadcastNodeStatus()
	go cluster.monitorCluster()

	return cluster, nil
}

func (c *Cluster) broadcastNodeStatus() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.nodesMutex.Lock()
			c.localNode.LastSeen = time.Now().Unix()
			c.nodesMutex.Unlock()

			c.updateNodeList()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Cluster) updateNodeList() {
	c.nodesMutex.Lock()
	defer c.nodesMutex.Unlock()

	for id, node := range c.nodes {
		if time.Now().Unix()-node.LastSeen > 30 {
			delete(c.nodes, id)
			log.Printf("Removed stale node: %s", id)
		}
	}

	for _, member := range c.mlist.Members() {
		if member.Name == c.config.InstanceID {
			continue
		}

		if _, exists := c.nodes[member.Name]; !exists {
			c.nodes[member.Name] = &NodeInfo{
				ID:       member.Name,
				HTTPAddr: fmt.Sprintf("http://%s:%d", member.Addr, c.config.HTTPPort),
				GRPCAddr: fmt.Sprintf("%s:%d", member.Addr, c.config.GRPCPort),
				LastSeen: time.Now().Unix(),
			}
			log.Printf("Discovered new node: %s", member.Name)
		}
	}
}

func (c *Cluster) monitorCluster() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Cluster members: %d", c.mlist.NumMembers())
		case <-c.stopChan:
			return
		}
	}
}

func (c *Cluster) GetNodes() []*NodeInfo {
	c.nodesMutex.RLock()
	defer c.nodesMutex.RUnlock()

	nodes := make([]*NodeInfo, 0, len(c.nodes)+1)
	nodes = append(nodes, c.localNode)
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (c *Cluster) Leave() error {
	close(c.stopChan)
	return c.mlist.Leave(time.Second * 10)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
