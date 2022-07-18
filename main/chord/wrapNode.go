package chord

import "math/big"

type WrapNode struct {
	node *Node
}

func (this *WrapNode) Getpredecessor(_ int, result *string) error {
	return this.node.get_predecessor(result)
}
func (this *WrapNode) SetSuccessorList(_ int, result *[successorListSize]string) error {
	return this.node.set_successor_list(result)
}
func (this *WrapNode) Notify(in string, _ *string) error {
	return this.node.notify(in)
}
func (this *WrapNode) FindSuccessor(in *big.Int, result *string) error {
	return this.node.find_successor(in, result)
}
func (this *WrapNode) CheckPredecessor(_ int, _ *string) error {
	return this.node.check_predecessor()
}
func (this *WrapNode) StoreData(dataPair Pair, _ *string) error {
	return this.node.store_data(dataPair)
}
func (this *WrapNode) StoreBackUp(dataPair Pair, _ *string) error {
	return this.node.store_backup(dataPair)
}
func (this *WrapNode) GetData(key string, value *string) error {
	return this.node.get_data(key, value)
}
func (this *WrapNode) DeleteData(key string, _ *string) error {
	return this.node.delete_data(key)
}
func (this *WrapNode) DeleteBackUp(key string, _ *string) error {
	return this.node.delete_backup(key)
}
func (this *WrapNode) HereditaryData(address string, dataset *map[string]string) error {
	return this.node.hereditary_data(address, dataset)
}
func (this *WrapNode) SubBackUp(dataset map[string]string, _ *string) error {
	return this.node.sub_backup(dataset)
}
func (this *WrapNode) AddBackUp(dataset map[string]string, _ *string) error {
	return this.node.add_backup(dataset)
}
func (this *WrapNode) GenerateBackUp(_ *string, dataset *map[string]string) error {
	return this.node.generate_backup(dataset)
}
func (this *WrapNode) Stabilize(_ int, _ *string) error {
	this.node.stabilize(true)
	return nil
}
func (this *WrapNode) Debug(in int, _ *string) error {
	this.node.debug(in)
	return nil
}
func (this *WrapNode) GetDebug(key string, value *string) error {
	this.node.get_debug(key, value)
	return nil
}
