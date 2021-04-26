package topology

import (
	"fmt"

	topo "github.com/kloudfuse/topology/gogen"
	"github.com/open-fresh/avalanche/metrics"
	"github.com/open-fresh/avalanche/utils"
)

type EntityNode struct {
	Cfg      *EntityConfig
	Entity   *topo.Entity
	Parent   *EntityNode
	Children []*EntityNode

	metrics []*metrics.Metric
}

func NewNode() *EntityNode {
	return &EntityNode{}
}

func (e *EntityNode) GetName() string {
	return e.Cfg.Name
}

func (e *EntityNode) SetParent(parent *EntityNode) *EntityNode {
	e.Parent = parent
	e.Parent.AddChild(e)
	return e
}

func (e *EntityNode) AddChild(child *EntityNode) *EntityNode {
	e.Children = append(e.Children, child)
	return e
}

func (e *EntityNode) Initialize(name string) {
	k := 0
	e.metrics = make([]*metrics.Metric, e.Cfg.MetricCount)

	for idx := 0; idx < e.Cfg.MetricCount; idx++ {
		name := fmt.Sprintf("%s_%v_%v", name, e.GetName(), idx)
		numLabels := e.Cfg.LabelCount
		var metric *metrics.Metric
		if e.Parent != nil {
			parentId := e.Parent.GetName() + "_id"
			metric = metrics.NewMetrics(name, numLabels, parentId)
		} else {
			metric = metrics.NewMetrics(name, numLabels, "")
		}
		metric.Register()
		e.metrics[k] = metric
		k++
	}
}

func (e *EntityNode) MakeEntity(entityIdStr string, parentId string) *topo.Entity {
	entityId := topo.EntityIdentifier_Uuid{Uuid: entityIdStr}
	entityMeta := topo.EntityIdentifier{Type: e.GetName(), Identifier: &entityId}

	entity := new(topo.Entity)
	entity.Metadata = &entityMeta
	entity.Operation = topo.Entity_UPDATE

	if e.Parent != nil {
		makeAttribute(e.Parent.GetName()+"_id", parentId)
	}

	for idx := 0; idx < e.Cfg.AttributeCount; idx++ {
		attr := makeAttribute("attr_"+utils.Convert(idx), "attr_val"+utils.Convert(idx))
		entity.Attributes = append(entity.Attributes, &attr)
	}

	return entity
}

func makeAttribute(key string, value string) topo.Attribute {
	return topo.Attribute{Key: key, Value: value}
}
