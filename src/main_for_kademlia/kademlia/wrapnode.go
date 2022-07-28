package kademlia

type WrapNode struct {
	node *Node
}

func (this *WrapNode) Ping(request Contact, _ *string) error {
	if this.node.Ping(request.Address) {
		this.node.table.Update(&request) ///
	}
	return nil
}
func (this *WrapNode) GetClose(input FindNodeRequest, result *FindNodeReply) error {
	result.Content = this.node.table.FindClosest(input.Target, K)
	result.Replier = this.node.addr
	result.Requester = input.Requester
	this.node.table.Update(&input.Requester)
	return nil
}
func (this *WrapNode) Store(input StoreRequest, _ *string) error {
	this.node.data.Store(input)
	this.node.table.Update(&input.Requester)
	return nil
}
func (this *WrapNode) FindValue(request FindValueRequest, reply *FindValueReply) error {
	reply.Content = this.node.table.FindClosest(Hash(request.Key), K)
	reply.Request = request.Request
	reply.Reply = this.node.addr
	reply.IsFind, reply.Value = this.node.data.get(request.Key)
	this.node.table.Update(&request.Request)
	return nil
}
