package topology

import (
	log "github.com/sirupsen/logrus"
)

type LoggingTopoVisitor struct{}

func (v LoggingTopoVisitor) visit(node *EntityNode, entityId string, parentId string) {
	e := node.MakeEntity(entityId, parentId)
	log.Debugf("Visiting %s with id %s", node.GetName(), e.String())
}
