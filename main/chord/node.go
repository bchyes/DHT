package chord

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
	"time"
)

const successorListSize int = 5
const hashBitsSize int = 160

type Node struct {
	address    string
	ID         *big.Int
	dataLock   sync.RWMutex
	dataSet    map[string]string
	QuitSignal chan bool
	backupLock sync.RWMutex
	backupSet  map[string]string
	rwLock     sync.RWMutex
	next       int
	station    *network

	successorList [successorListSize]string
	predecessor   string
	fingerTable   [hashBitsSize]string
}

func (this *Node) Init(port int) {
	this.address = fmt.Sprintf("%s:%d", localAddress, port)
	this.ID = ConsistentHash(this.address)
	//
	this.reset()
}
func (this *Node) Run() {
	this.station = new(network)
	err := this.station.Init(this.address, this)
	if err != nil {
		log.Errorln("<Run> failed ", err)
		return
	}
	log.Infoln("<Run> success in", this.address)
	//
	this.next = 1
}
func (this *Node) Create() {
	this.predecessor = ""
	this.successorList[0] = this.address
	this.fingerTable[0] = this.address
	this.background()
	log.Infoln("<Create> new ring success in ", this.address)
}
func (this *Node) Join(addr string) bool {
	return false
}
func (this *Node) Quit() {

}
func (this *Node) ForceQuit() {

}
func (this *Node) Ping(addr string) bool {
	return false
}
func (this *Node) Put(key string, value string) bool {
	return false
}
func (this *Node) Get(key string) (bool, string) {
	return false, ""
}
func (this *Node) Delete(key string) bool {
	return false
}
func (this *Node) first_online_successor() (error, string) {
	for i := 0; i < successorListSize; i++ {
		if CheckOnline(this.successorList[i]) {
			return nil, this.successorList[i]
		}
	}
	log.Errorln("<first_online_successor> List Break in ", this.address)
	return errors.New(fmt.Sprintln("<first_online_successor> List Break in ", this.address)), ""
}
func (this *Node) stabilize() {
	_, newSucAddr := this.first_online_successor()
	var succPredAddr string
	err := RemoteCall(newSucAddr, "wrapNode.Getpredecessor", 2022, &succPredAddr)
	if err != nil {

	}
}
func (this *Node) background() {
	go func() {
		if this.conRoutineFlag {
			this.stabilize()
			time.Sleep(timeCut)
		}
	}()
	go func() {
		if this.conRoutineFlag {
			this.check_predecessor()
			time.Sleep(timeCut)
		}
	}()
	go func() {
		if this.comRoutineFlag {
			this.fix_finger()
			time.Sleep(timeCut)
		}
	}()
}
func (this *Node) get_predecessor(result *string) error {

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
