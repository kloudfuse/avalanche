package topology

import (
	log "github.com/sirupsen/logrus"
)

type Topology struct {
	Name  string
	roots []*EntityNode
	nodes map[string]*EntityNode
}

func NewTopology(name string) *Topology {
	t := new(Topology)
	t.Name = name
	t.roots = make([]*EntityNode, 0)
	t.nodes = make(map[string]*EntityNode)
	return t
}

func (t *Topology) AddNode(cfg *EntityConfig) *EntityNode {
	if existingNode, ok := t.nodes[cfg.Name]; ok {
		t.updateNode(cfg, existingNode)
		return existingNode
	}

	return t.MakeNode(cfg)
}

func (t *Topology) MakeNode(cfg *EntityConfig) *EntityNode {
	node := NewNode()
	node.Cfg = cfg

	t.updateParent(cfg.Parent, node)
	t.AddEntity(node)
	return node
}

func (t *Topology) AddEntity(node *EntityNode) {
	t.nodes[node.GetName()] = node
}

func (t *Topology) updateNode(cfg *EntityConfig, node *EntityNode) {
	node.Cfg = cfg
	t.updateParent(cfg.Parent, node)
}

func (t *Topology) updateParent(parentName string, node *EntityNode) {
	if parentName == "" {
		t.roots = append(t.roots, node)
		return
	}

	if _, ok := t.nodes[parentName]; !ok {
		newParent := new(EntityNode)
		t.nodes[parentName] = newParent
	}

	node.SetParent(t.nodes[parentName])
}

func (t *Topology) Intialize() {
	for _, root := range t.roots {
		t.InitializeNode(root)
	}
}

func (t *Topology) InitializeNode(node *EntityNode) {
	if node == nil {
		return
	}
	if node.Parent == nil {
		log.Infof("Root %s", node.GetName())
	} else {
		log.Infof("%s", node.GetName())
	}
	node.Initialize(t.Name)
	for _, child := range node.Children {
		t.InitializeNode(child)
	}
}
