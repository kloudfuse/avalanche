package topology

type MetricTopoVisitor struct {
	cycleId int
}

func (v MetricTopoVisitor) visit(node *EntityNode, entityId string, parentId string) {
	v.cycleId++
	for _, metric := range node.metrics {
		metric.Publish(v.cycleId, parentId)
	}
}
