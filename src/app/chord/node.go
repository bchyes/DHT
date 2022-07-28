package chord

import (
	"errors"
	"fmt"
	//log "github.com/sirupsen///logrus"
	"math/big"
	"sync"
	"time"
	//"time"
)

const successorListSize int = 5
const hashBitsSize int = 160

var wg sync.WaitGroup

type Node struct {
	address        string
	ID             *big.Int
	dataLock       sync.RWMutex
	dataSet        map[string]string
	QuitSignal     chan bool
	backupLock     sync.RWMutex
	backupSet      map[string]string
	rwLock         sync.RWMutex
	next           int
	station        *network
	conRoutineFlag bool

	successorList [successorListSize]string
	predecessor   string
	fingerTable   [hashBitsSize]string
}

func (this *Node) Init(port int) {
	this.address = fmt.Sprintf("%s:%d", localAddress, port)
	this.ID = ConsistentHash(this.address)
	this.conRoutineFlag = false
	this.reset()
}
func (this *Node) Run() {
	this.station = new(network)
	err := this.station.Init(this.address, this)
	if err != nil {
		//log.Errorln("<Run> failed ", err)
		return
	}
	//log.Infoln("<Run> success in", this.address)
	this.conRoutineFlag = true
	this.next = 1 //
}
func (this *Node) Create() {
	this.rwLock.Lock()
	this.predecessor = ""
	this.successorList[0] = this.address
	this.fingerTable[0] = this.address
	this.rwLock.Unlock()
	this.background()
	//log.Infoln("<Create> new ring success in ", this.address)
}
func (this *Node) Join(addr string, j int) bool {
	if !CheckOnline(addr) {
		//log.Errorln("<Join> fail in ", addr, "because this node is not online")
		return false
	}
	var succAddr string
	err := RemoteCall(addr, "WrapNode.FindSuccessor", this.ID, &succAddr)
	if err != nil {
		//log.Errorln("<Join> fail in find successor in ", addr)
		return false
	}
	var result [successorListSize]string
	err = RemoteCall(succAddr, "WrapNode.SetSuccessorList", 2022, &result)
	if err != nil {
		//log.Errorln("<Join> fail in find successorList in ", succAddr)
		return false
	}
	this.rwLock.Lock()
	this.predecessor = "" //!
	this.successorList[0] = succAddr
	this.fingerTable[0] = succAddr
	for i := 1; i < successorListSize; i++ {
		this.successorList[i] = result[i-1]
	}
	this.rwLock.Unlock()
	this.dataLock.Lock()
	err = RemoteCall(succAddr, "WrapNode.HereditaryData", this.address, &this.dataSet)
	this.dataLock.Unlock()
	if err != nil {
		//log.Errorln("<Join> fail in HereditaryData in ", succAddr)
		return false
	}
	this.background()
	//log.Infoln("<Join> success in ", this.address)
	//if j >= 0 {
	//	time.Sleep(20 * timeCut)
	//	fmt.Print(this.address)
	//	for i := 0; i < successorListSize; i++ {
	//		fmt.Print("->", this.successorList[i])
	//	}
	//	fmt.Print("\n")
	//	var occupy string
	//	var result string
	//	_, result = this.first_online_successor()
	//	_ = RemoteCall(result, "WrapNode.Debug", 1, &occupy)
	//	//time.Sleep(timeCut)
	//	//wg.Wait()
	//}
	return true
}
func (this *Node) Quit(i int) {
	if !CheckOnline(this.address) {
		return
	}
	//if i == 1 {
	//	fmt.Print("FirstQuit\n")
	//	time.Sleep(20 * timeCut)
	//	fmt.Print(this.predecessor)
	//	var occupy string
	//	var result string
	//	_, result = this.first_online_successor()
	//	_ = RemoteCall(result, "WrapNode.Debug", 1, &occupy)
	//	//time.Sleep(timeCut)
	//	//wg.Wait()
	//}
	this.rwLock.Lock()
	this.conRoutineFlag = false
	this.rwLock.Unlock()
	time.Sleep(3 * waitTime)
	err := this.station.ShutDown()
	if err != nil {
		//log.Errorln("<Quit> fail in shutdown in ", this.address)
		return
	}
	////log.Warningln("QUIT SHUTDOWN in", this.address)
	//var result string
	//err, result = this.first_online_successor()
	//if err != nil {
	//	//log.Errorln("<Quit> fail in get FirstOnlineSuccessor in ", this.address)
	//	return
	//}
	//var occupy string
	//err = RemoteCall(result, "WrapNode.CheckPredecessor", 2022, &occupy)
	//if err != nil {
	//	//log.Errorln("<Quit> fail in CheckPredecessor in ", result)
	//	return
	//}
	var occupy string
	//log.Warningln("QUIT SHUTDOWN in", this.address, " and predecessor is ", this.predecessor)
	fmt.Print("QUIT SHUTDOWN in ", this.address, " and predecessor is ", this.predecessor)
	fmt.Print("\n")
	err = RemoteCall(this.predecessor, "WrapNode.Stabilize", 2022, &occupy)
	if err != nil {
		//log.Errorln("<Quit> fail in Stabilize")
		return
	}
	//log.Infoln("<Quit> success in ", this.address)
	fmt.Print("<Quit> success in ", this.address)
	fmt.Print("\n")
	////log.Warningln("QUIT STABILIZE in ", this.address)
	this.reset()
	//if i >= 0 {
	//	time.Sleep(20 * timeCut)
	//	fmt.Print(this.predecessor)
	//	var occupy string
	//	var result string
	//	_, result = this.first_online_successor()
	//	_ = RemoteCall(result, "WrapNode.Debug", 1, &occupy)
	//	//time.Sleep(timeCut)
	//	//wg.Wait()
	//}
}
func (this *Node) ForceQuit() {
	if !CheckOnline(this.address) {
		return
	}
	this.rwLock.Lock()
	this.conRoutineFlag = false
	this.rwLock.Unlock()
	time.Sleep(3 * timeCut)
	err := this.station.ShutDown()
	if err != nil {
		//log.Errorln("<ForceQuit> fail in ShutDown in ", this.address)
		return
	}
	this.reset()
	//log.Infoln("<ForceQuit> success in ", this.address)
}
func (this *Node) Ping(addr string) bool {
	return CheckOnline(addr)
}
func (this *Node) Put(key string, value string) bool {
	if !this.conRoutineFlag {
		//log.Errorln("<Put> fail in ", this.address, "because the node is sleep")
		return false
	}
	var target string
	err := this.find_successor(ConsistentHash(key), &target)
	if err != nil {
		//log.Errorln("<Put> fail find successor in ", this.address)
		return false
	}
	var occupy string
	var dataPair = Pair{key, value}
	err = RemoteCall(target, "WrapNode.StoreData", dataPair, &occupy)
	if err != nil {
		//log.Errorln("<Put> fail StoreDate in ", this.address)
		return false
	}
	//log.Infoln("<Put> success in ", this.address)
	return true
}
func (this *Node) Get(key string) (bool, string) {
	if !this.conRoutineFlag {
		//log.Errorln("<Get> fail in ", this.address, "because the node is sleep")
		return false, ""
	}
	var result string
	err := this.find_successor(ConsistentHash(key), &result)
	if err != nil {
		//log.Errorln("<Get> fail find successor in ", this.address)
		return false, ""
	}
	var value string
	err = RemoteCall(result, "WrapNode.GetData", key, &value)
	if err != nil {
		//log.Errorln("<Get> fail GetData in ", result)
		fmt.Print("<Get> fail GetData in ", result, "\n")
		_ = RemoteCall(result, "WrapNode.GetDebug", key, &value)
		return false, ""
	}
	//log.Infoln("<Get> success in ", this.address)
	return true, value
}
func (this *Node) Delete(key string) bool {
	if !this.conRoutineFlag {
		//log.Errorln("<Delete> fail in ", this.address, "because the node is sleep")
		return false
	}
	var result string
	err := this.find_successor(ConsistentHash(key), &result)
	if err != nil {
		//log.Errorln("<Delete> fail in find successor in ", this.address)
		return false
	}
	var occupy string
	err = RemoteCall(result, "WrapNode.DeleteData", key, &occupy)
	if err != nil {
		//log.Errorln("<Delete> fail in delete node in ", result)
		return false
	}
	//log.Infoln("<Delete> success in delete node in ", result)
	return true
}
func (this *Node) find_successor(target *big.Int, result *string) error {
	err, succAddr := this.first_online_successor()
	if err != nil {
		//log.Errorln("<Find Successor> Failed because", err)
	}
	if Contain(target, this.ID, ConsistentHash(succAddr), true) {
		*result = succAddr
		return nil
	}
	closestPre := this.closest_preceding_finger(target)
	return RemoteCall(closestPre, "WrapNode.FindSuccessor", target, result)
}
func (this *Node) closest_preceding_finger(target *big.Int) string {
	for i := hashBitsSize - 1; i >= 0; i-- {
		if this.fingerTable[i] == "" {
			continue
		}
		if !CheckOnline(this.fingerTable[i]) {
			continue
		}
		if Contain(ConsistentHash(this.fingerTable[i]), this.ID, target, false) {
			//log.Infoln("<ClosestPrecedingFinger> success in Node", this.address)
			return this.fingerTable[i]
		}
	}
	err, preaddr := this.first_online_successor()
	if err != nil {
		//log.Errorln("<ClosestPrecedingFinger> Failed in find FirstOnlineSuccessor in ", this.address)
		return ""
	}
	return preaddr
}
func (this *Node) first_online_successor() (error, string) {
	for i := 0; i < successorListSize; i++ {
		if CheckOnline(this.successorList[i]) {
			return nil, this.successorList[i]
		}
	}
	//log.Errorln("<first_online_successor> List Break in ", this.address)
	return errors.New(fmt.Sprintln("<first_online_successor> List Break in ", this.address)), ""
}
func (this *Node) stabilize(flag bool) {
	if flag {
		//	//log.Warningln("st1")
		//log.Warningln("stabilize address in ", this.address)
	}
	_, newSucAddr := this.first_online_successor()
	var occupy string
	if flag {
		//	//log.Warningln("st2")
		//log.Warningln(newSucAddr)
	}
	err := RemoteCall(newSucAddr, "WrapNode.CheckPredecessor", 2022, &occupy)
	if err != nil {
		//log.Errorln("<Stabilize> fail in CheckPredecessor in ", newSucAddr)
		return
	}
	var succPredAddr string
	//if flag {
	//	//log.Warningln("st3")
	//}
	err = RemoteCall(newSucAddr, "WrapNode.Getpredecessor", 2022, &succPredAddr)
	if err != nil {
		//log.Errorln("<Stabilize> Fail to get predecessor in ", newSucAddr)
		return
	}
	//if flag {
	//	//log.Warningln("st4")
	//}
	if succPredAddr != "" && Contain(ConsistentHash(succPredAddr), this.ID, ConsistentHash(newSucAddr), false) {
		newSucAddr = succPredAddr
	}
	var tmp [successorListSize]string
	//if flag {
	//	//log.Warningln("st5")
	//}
	err = RemoteCall(newSucAddr, "WrapNode.SetSuccessorList", 2022, &tmp)
	if err != nil {
		//log.Errorln("<Stabilize> Fail to stabilize,can't get SuccessorList in", newSucAddr)
	}
	//if flag {
	//	//log.Warningln("st6")
	//}
	this.rwLock.Lock()
	this.successorList[0] = newSucAddr
	this.fingerTable[0] = newSucAddr
	for i := 1; i < successorListSize; i++ {
		this.successorList[i] = tmp[i-1]
	}
	this.rwLock.Unlock()
	//var occupy string
	err = RemoteCall(newSucAddr, "WrapNode.Notify", this.address, &occupy)
	if err != nil {
		//log.Errorln("<Stabilize> Fail to Notify in ", newSucAddr, "because ", err)
	}
	//if flag {
	//	//log.Warningln("st7")
	//}
	//log.Infoln("<Stabilize> Success in ", this.address)
}
func (this *Node) notify(in string) error {
	if this.predecessor == in {
		return nil
	}
	if this.predecessor == "" || Contain(ConsistentHash(in), ConsistentHash(this.predecessor), this.ID, false) {
		//log.Warningln("notify prepreodecessor is ", this.predecessor)
		this.rwLock.Lock()
		this.predecessor = in
		this.rwLock.Unlock()
		var occupy string
		this.backupLock.Lock()
		err := RemoteCall(in, "WrapNode.GenerateBackUp", &occupy, &this.backupSet)
		this.backupLock.Unlock()
		if err != nil {
			//log.Errorln("<Notify> fail in GenerateBackup in ", in)
			return err
		}
		//log.Infoln("<Notify> change ", this.address, " Predecessor to", this.predecessor)
	}
	return nil
}
func (this *Node) check_predecessor() error {
	////log.Infoln(this.address, "->", this.predecessor)
	if this.predecessor != "" && !CheckOnline(this.predecessor) {
		this.rwLock.Lock()
		this.predecessor = ""
		this.rwLock.Unlock()
		////log.Warningln(this.address, "->", this.predecessor)
		this.apply_backup()
		////log.Warningln(this.address, "->", this.predecessor)
		//log.Infoln("<CheckPredecessor> fail in find predecossor in ", this.address)
	}
	return nil
}
func (this *Node) fix_finger() {
	var result string
	err := this.find_successor(CalculateID(this.ID, this.next), &result)
	if err != nil {
		//log.Errorln("<FixFinger> Fail in find successor in ", this.address, "+ next = ", this.next)
		return
	}
	this.rwLock.Lock()
	this.fingerTable[0] = this.successorList[0]
	this.fingerTable[this.next] = result
	this.next++
	if this.next >= hashBitsSize {
		this.next = 1
	}
	this.rwLock.Unlock()
	//log.Infoln("<FixFinger> success in ", this.address)
}
func (this *Node) background() {
	go func() {
		for this.conRoutineFlag {
			this.stabilize(false)
			time.Sleep(timeCut)
		}
	}()
	//go func() {
	//	for this.conRoutineFlag {
	//		this.check_predecessor() //!
	//		time.Sleep(timeCut)
	//	}
	//}()
	go func() {
		for this.conRoutineFlag {
			this.fix_finger()
			time.Sleep(timeCut)
		}
	}()
}
func (this *Node) get_predecessor(result *string) error {
	this.rwLock.RLock()
	*result = this.predecessor
	this.rwLock.RUnlock()
	return nil
}
func (this *Node) set_successor_list(result *[successorListSize]string) error {
	this.rwLock.RLock()
	*result = this.successorList
	this.rwLock.RUnlock()
	return nil
}
func (this *Node) store_data(dataPair Pair) error {
	this.dataLock.Lock()
	this.dataSet[dataPair.Key] = dataPair.Value
	this.dataLock.Unlock()
	err, result := this.first_online_successor()
	if err != nil {
		//log.Errorln("<StoreData> fail in find FirstOnlineSuccessor in ", this.address)
		return err
	}
	var occupy string
	err = RemoteCall(result, "WrapNode.StoreBackUp", dataPair, &occupy)
	if err != nil {
		//log.Errorln("<StoreData> fail in StoreBackUp in ", result)
		return err
	}
	return nil
}
func (this *Node) get_data(key string, value *string) error {
	var ok bool
	this.dataLock.RLock()
	*value, ok = this.dataSet[key]
	this.dataLock.RUnlock()
	if !ok {
		*value = ""
		//log.Errorln("<GetData> fail in get ", key)
		return errors.New(fmt.Sprintln("<GetData> fail in get", key))
	} else {
		//log.Infoln("<GetData> success in get ", key)
		return nil
	}
}
func (this *Node) delete_data(key string) error {
	this.dataLock.Lock()
	_, ok := this.dataSet[key]
	if ok {
		delete(this.dataSet, key)
	}
	this.dataLock.Unlock()
	if ok {
		err, result := this.first_online_successor()
		if err != nil {
			//log.Errorln("<DeleteData> fail in find FirstOnlineSuccessor in ", this.address)
			return err
		}
		var occupy string
		err = RemoteCall(result, "WrapNode.DeleteBackUp", key, &occupy)
		if err != nil {
			//log.Errorln("<DeleteData> fail in DeleteBackUp in ", result)
			return err
		}
		return nil
	} else {
		return errors.New(fmt.Sprintln("<DeleteData> fail in delete ", key, " at ", this.address))
	}
}
func (this *Node) hereditary_data(preaddress string, dataset *map[string]string) error {
	this.backupLock.Lock()
	this.dataLock.Lock()
	this.backupSet = make(map[string]string) ////////////////
	for k, v := range this.dataSet {
		if !Contain(ConsistentHash(k), ConsistentHash(preaddress), this.ID, true) {
			(*dataset)[k] = v
			this.backupSet[k] = v
			delete(this.dataSet, k)
		}
	}
	this.backupLock.Unlock()
	this.dataLock.Unlock()
	err, result := this.first_online_successor()
	if err != nil {
		//log.Errorln("<HereditaryData> fail in find FirstOnlineSuccessor in ", this.address)
		return err
	}
	var occupy string
	err = RemoteCall(result, "WrapNode.SubBackUp", *dataset, &occupy)
	if err != nil {
		//log.Errorln("<HereditaryData> fail in SubBackUp in ", result)
		return err
	}
	this.rwLock.Lock()
	this.predecessor = preaddress
	this.rwLock.Unlock()
	//log.Infoln("<HereditaryData> success in ", this.address)
	return nil
}
func (this *Node) sub_backup(dataset map[string]string) error {
	this.backupLock.Lock()
	for k := range dataset {
		delete(this.backupSet, k)
	}
	this.backupLock.Unlock()
	return nil
}
func (this *Node) store_backup(dataPair Pair) error {
	this.backupLock.Lock()
	this.backupSet[dataPair.Key] = dataPair.Value
	this.backupLock.Unlock()
	return nil
}
func (this *Node) delete_backup(key string) error {
	this.backupLock.Lock()
	_, ok := this.backupSet[key]
	if ok {
		delete(this.backupSet, key)
	}
	this.backupLock.Unlock()
	if ok {
		return nil
	} else {
		return errors.New(fmt.Sprintln("<DeleteBackUp> fail in delete ", key, " at ", this.address))
	}
}
func (this *Node) add_backup(dataset map[string]string) error {
	this.backupLock.Lock()
	for k, v := range dataset {
		this.backupSet[k] = v
	}
	this.backupLock.Unlock()
	return nil
}
func (this *Node) generate_backup(dataset *map[string]string) error {
	//fmt.Println("GenerateBackUp")
	this.dataLock.RLock()
	*dataset = make(map[string]string)
	for k, v := range this.dataSet {
		(*dataset)[k] = v
		//fmt.Println(k, " ", v)
	}
	this.dataLock.RUnlock()
	return nil
}
func (this *Node) apply_backup() error {
	this.backupLock.RLock()
	this.dataLock.Lock()
	for k, v := range this.backupSet {
		this.dataSet[k] = v
	}
	this.backupLock.RUnlock() //////
	this.dataLock.Unlock()    //////
	err, result := this.first_online_successor()
	if err != nil {
		//log.Errorln("<ApplyBackUp> fail in find FirstOnlineSuccessor in ", this.address)
		return err
	}
	var occupy string
	////log.Warningln("<ApplyBackUp>")
	err = RemoteCall(result, "WrapNode.AddBackUp", this.backupSet, &occupy)
	////log.Warningln("<ApplyBackUp>")
	if err != nil {
		//log.Errorln("<ApplyBackUp> fail in AddBackUp in ", result)
		return err
	}
	////log.Warningln("<start BackUpLock>")
	this.backupLock.Lock()
	////log.Warningln("<BackUpLocking>")
	this.backupSet = make(map[string]string)
	this.backupLock.Unlock()
	//log.Infoln("<ApplyBackUp> success in ", this.address)
	return nil
}
func (this *Node) reset() { //
	this.dataLock.Lock() //
	this.dataSet = make(map[string]string)
	this.dataLock.Unlock() //
	this.QuitSignal = make(chan bool, 2)
	this.backupLock.Lock() //
	this.backupSet = make(map[string]string)
	this.backupLock.Unlock() //
	this.rwLock.Lock()       //
	this.next = 1
	this.rwLock.Unlock() //
}
func (this *Node) debug(in int) {
	//wg.Add(1)
	fmt.Print("->", this.address)
	//log.Infoln("->", this.address)
	for i := 0; i < successorListSize; i++ {
		fmt.Print("->", this.successorList[i])
		//log.Infoln("->", this.successorList[i])
	}
	fmt.Print("\n")
	fmt.Print("dataset\n")
	//log.Infoln("dataset\n")
	for k, v := range this.dataSet {
		//log.Infoln(k, " ", v, "\n")
		fmt.Print(k, " ", v, "\n")
	}
	fmt.Print("backup\n")
	//log.Infoln("backup\n")
	for k, v := range this.backupSet {
		//log.Infoln(k, " ", v, "\n")
		fmt.Print(k, " ", v, "\n")
	}
	if in == 21 {
		return
	}
	var result string
	var occupy string
	_, result = this.first_online_successor()
	//defer wg.Done()
	_ = RemoteCall(result, "WrapNode.Debug", in+1, &occupy)
}
func (this *Node) get_debug(key string, value *string) {
	err := this.get_data(key, value)
	if err == nil {
		return
	} else {
		fmt.Print("<GetDebug> fail GetData in ", this.address, "\n")
		//_, result := this.first_online_successor()
		//_ = RemoteCall(result, "WrapNode.GetDebug", key, value)
	}
}
