package topology

import (
	"github.com/open-fresh/avalanche/utils"
)

type TopoVisitor interface {
	visit(node *EntityNode, entityId string, parentId string)
}

type TopoWalker struct {
	tp       *Topology
	visitors []TopoVisitor
}

func NewWalker(tp *Topology) *TopoWalker {
	visitors := make([]TopoVisitor, 0)
	return &TopoWalker{tp, visitors}
}

func (tw *TopoWalker) AddVisitor(visitor TopoVisitor) {
	tw.visitors = append(tw.visitors, visitor)
}

func (tw *TopoWalker) Walk() {
	for _, root := range tw.tp.roots {
		tw.WalkEntity(root, "")
	}
}

func (tw *TopoWalker) WalkEntity(node *EntityNode, parentId string) {
	for idx := 0; idx < node.Cfg.Count; idx++ {
		entityId := node.GetName() + "-" + utils.Convert(idx)

		for _, visitor := range tw.visitors {
			visitor.visit(node, entityId, parentId)
		}
		//e := node.MakeEntity(entityId, parentId)

		for _, child := range node.Children {
			tw.WalkEntity(child, entityId)
		}
	}
}
