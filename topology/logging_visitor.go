package topology

import (
	log "github.com/sirupsen/logrus"
)

type LoggingTopoVisitor struct {
	Name string
}

func (v LoggingTopoVisitor) visit(node *EntityNode, entityId string, parentId string) {
	log.Infof("%s: Visiting node %s with id %s and parent %s", v.Name, node.GetName(), entityId, parentId)
}
