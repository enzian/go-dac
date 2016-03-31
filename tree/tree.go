package tree

type TreeNode struct {
	ID       interface{}
	Depth    int64
	children []*TreeNode
}

func (tn *TreeNode) AppendChild(id interface{}) {
	var newNode = TreeNode{ID: id, Depth: tn.Depth + 1}
	tn.children = append(tn.children, &newNode)
}

func (tn *TreeNode) IsLeafNode() bool {
	return len(tn.children) == 0
}

func (tn *TreeNode) Flatten() map[interface{}]int64 {
	var nodeBacklog = []*TreeNode{tn}
	var flatList = make(map[interface{}]int64)

	for len(nodeBacklog) > 0 {
		var currentNode = nodeBacklog[0]

		if currentNode.IsLeafNode() {
			nodeBacklog = nodeBacklog[1:]
		} else {
			nodeBacklog = append(currentNode.children, nodeBacklog[1:]...)
		}

		flatList[currentNode.ID] = currentNode.Depth
	}

	return flatList
}
