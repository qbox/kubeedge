package cloudidmanager

import (
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/klog/v2"
)

type Manager struct {
	NodeNumber int32

	Nodes sync.Map
}

type NodeConnectionInfo struct {
	createAt time.Time
	nodeID   string
	cloudID  string
}

func NewNodeConnectionInfo(nodeID, cloudID string) *NodeConnectionInfo {
	return &NodeConnectionInfo{
		createAt: time.Now(),
		nodeID:   nodeID,
		cloudID:  cloudID,
	}
}

func NewCloudConnectionManager() *Manager {
	return &Manager{
		Nodes: sync.Map{},
	}
}

func (ccm *Manager) AddNode(nodeConn *NodeConnectionInfo) {
	nodeID := nodeConn.nodeID

	ons, exist := ccm.Nodes.LoadAndDelete(nodeID)
	if exist {
		if oldNodeID, ok := ons.(NodeConnectionInfo); ok {
			klog.Warningf("node exists for %s", oldNodeID)
			atomic.AddInt32(&ccm.NodeNumber, -1)
		}
	}

	ccm.Nodes.Store(nodeID, nodeConn)
	atomic.AddInt32(&ccm.NodeNumber, 1)
}

func (ccm *Manager) DeleteNode(nodeConn *NodeConnectionInfo) {
	cacheNode, exist := ccm.GetCloud(nodeConn.nodeID)
	if !exist {
		klog.Warningf("node %s not found", nodeConn.nodeID)
		return
	}

	// do this actually need?
	if cacheNode != nodeConn {
		klog.Warningf("node %s not found", nodeConn.nodeID)
		return
	}

	ccm.Nodes.Delete(nodeConn.nodeID)
	atomic.AddInt32(&ccm.NodeNumber, -1)
}

func (ccm *Manager) GetCloud(nodeID string) (*NodeConnectionInfo, bool) {
	ons, exists := ccm.Nodes.Load(nodeID)
	if exists {
		return ons.(*NodeConnectionInfo), true
	}

	return nil, false
}