package chord

type wrapNode struct {
	node *Node
}

func (this *wrapNode) Getpredecessor(_ int, result *string) error {
	return this.node.get_predecessor(result)
}
