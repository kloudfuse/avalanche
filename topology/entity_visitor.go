package topology

type PgTopoVisitor struct {
}

func (v PgTopoVisitor) visit(node *EntityNode, entityId string, parentId string) {
	_ = node.MakeEntity(entityId, parentId)
	// write to postgres here
}
