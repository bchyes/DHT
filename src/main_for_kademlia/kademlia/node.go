package kademlia

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

const IDlength = 20

type ID [IDlength]byte

const K = 100
const alpha = 3
const backgroundInterval1 = 5 * time.Second
const backgroundInterval2 = 10 * time.Minute
const refreshTimeInterval = 30 * time.Second

type Contact struct {
	Address string
	NodeID  ID
}
type ContactRecord struct {
	SortKey     ID
	ContactInfo Contact
}
type Node struct {
	QuitSignal chan bool
	addr       Contact
	station    *network
	isRunning  bool
	data       database
	table      RoutingTable
}
type FindNodeReply struct {
	Requester Contact
	Replier   Contact
	Content   []ContactRecord
}
type FindNodeRequest struct {
	Requester Contact
	Target    ID
}
type StoreRequest struct {
	Key          string
	Value        string
	RequesterPri int
	Requester    Contact
}
type FindValueRequest struct {
	Key     string
	Request Contact
}
type FindValueReply struct {
	Request Contact
	Reply   Contact
	Content []ContactRecord
	IsFind  bool
	Value   string
}

func (this *Node) reset() {
	this.QuitSignal = make(chan bool, 2)
	this.isRunning = false
	this.table.InitRoutingTable(this.addr)
	this.data.Init()
}
func (this *Node) Init(port int) {
	this.addr = NewContact(fmt.Sprintf("%s:%d", localAddress, port))
	this.reset()
}
func (this *Node) Run() {
	//log.Warningln(this.addr.Address)
	this.station = new(network)
	err := this.station.Init(this.addr.Address, this)
	if err != nil {
		log.Errorln("<Run> error in Init in ", this.addr.Address)
		return
	}
	//log.Infoln("<Run> success in ", this.addr.Address)
	this.isRunning = true
	this.background()
}
func (this *Node) background() {
	go func() {
		for this.isRunning {
			//
			time.Sleep(backgroundInterval1)
		}
	}()
	go func() {
		for this.isRunning {
			this.Republic()
			time.Sleep(backgroundInterval2)
		}
	}()
	go func() {
		for this.isRunning {
			this.Duplicate()
			time.Sleep(backgroundInterval2)
		}
	}()
	go func() {
		for this.isRunning {
			this.Expire()
			time.Sleep(backgroundInterval2)
		}
	}()
}
func (this *Node) Refresh() {
	lastRefreshTime := this.table.refreshTimeSet[this.table.refreshIndex]
	if !lastRefreshTime.Add(refreshTimeInterval).After(time.Now()) {
		tmpID := this.GenerateID(this.addr.NodeID, this.table.refreshIndex) /////////
		this.FindClosetNode(tmpID)
		this.table.refreshTimeSet[this.table.refreshIndex] = time.Now()
	}
	this.table.refreshIndex = (this.table.refreshIndex + 1) % (IDlength * 8)
}
func (this *Node) RangePut(request StoreRequest) {
	pendingList := this.FindClosetNode(Hash(request.Key))
	//for i := 0; i < len(pendingList); i++ {
	//	println(pendingList[i].ContactInfo.Address)
	//}
	count := new(int32)
	*count = 0
	for index := 0; index < len(pendingList); {
		if *count < alpha {
			target := pendingList[index].ContactInfo
			index++
			atomic.AddInt32(count, 1)
			go func(targetNode *Contact, input StoreRequest) {
				var occupy string
				err := RemoteCall(this, targetNode, "WrapNode.Store", input, &occupy)
				if err != nil {
					log.Errorln("<RangePut> fail in ", targetNode.Address, " because ", err)
				}
				atomic.AddInt32(count, -1)
			}(&target, request)
		} else {
			time.Sleep(sleepTime)
		}
	}
}
func (this *Node) Republic() {
	pendingList := this.data.republic()
	for k, v := range pendingList {
		request := StoreRequest{k, v, publisher, this.addr}
		this.RangePut(request)
	}
}
func (this *Node) Duplicate() {
	pendingList := this.data.duplicate()
	for k, v := range pendingList {
		request := StoreRequest{k, v, duplicater, this.addr}
		this.RangePut(request)
	}
}
func (this *Node) Expire() {
	this.data.expire()
}
func (this *Node) Create() {

}
func (this *Node) Join(addr string) bool {
	bootstrap := new(Contact)
	*bootstrap = NewContact(addr)
	if isOnline := CheckOnline(this, bootstrap); !isOnline {
		log.Errorln("<Join> fail in CheckOnline in ", bootstrap)
		return false
	}
	this.table.Update(bootstrap)
	this.FindClosetNode(this.addr.NodeID) //////cached
	return true
}
func (this *Node) FindClosetNode(target ID) []ContactRecord {
	resultList := make([]ContactRecord, 0, K*2)
	pendinglist := this.table.FindClosest(target, K)
	//for i := 0; i < len(pendinglist); i++ {
	//	println(pendinglist[i].ContactInfo.Address)
	//}
	//println("FindClosetNode")
	inRun := new(int32)
	*inRun = 0
	visit := make(map[string]bool)
	visit[this.addr.Address] = true
	index := 0
	ch := make(chan FindNodeReply, alpha+3)
	for index < len(pendinglist) || *inRun > 0 {
		for index < len(pendinglist) && *inRun < alpha {
			tmpReplier := pendinglist[index].ContactInfo
			if _, ok := visit[tmpReplier.Address]; !ok {
				visit[tmpReplier.Address] = true
				atomic.AddInt32(inRun, 1) //////
				go func(Replier *Contact, channel chan FindNodeReply) {
					var response FindNodeReply
					err := RemoteCall(this, Replier, "WrapNode.GetClose", FindNodeRequest{this.addr, target}, &response)
					if err != nil {
						atomic.AddInt32(inRun, -1)
						log.Errorln("<FindClosetNode> fail in GetClose due to ", err)
						return
					}
					ch <- response
					return
				}(&tmpReplier, ch)
			}
			index++
		}
		if *inRun > 0 {
			select {
			case response := <-ch:
				atomic.AddInt32(inRun, -1)
				resultList = append(resultList, ContactRecord{Xor(response.Replier.NodeID, target), response.Replier})
				for _, v := range response.Content {
					pendinglist = append(pendinglist, v)
				}
				SliceSort(&pendinglist)
			case <-time.After(waitTime):
				log.Warningln("<FindClosetNode> time out")
			}
		}
	}
	SliceSort(&resultList)
	if len(resultList) > K {
		resultList = resultList[:K]
	}
	return resultList
}
func (this *Node) Quit() {
	//log.Warningln("ShutDown0")
	//println("Quit address ", this.addr.Address)
	if !PureCheckConn(this.addr.Address) {
		return
	}
	err := this.station.ShutDown()
	if err != nil {
		log.Errorln("<Quit> fail in ", this.addr.Address)
	}
	this.reset()
}
func (this *Node) ForceQuit() {
	if !PureCheckConn(this.addr.Address) {
		return
	}
	_ = this.station.ShutDown()
	this.reset()
}
func (this *Node) Ping(addr string) bool {
	if !PureCheckConn(addr) {
		return false
	} else {
		return true
	}
}
func (this *Node) Put(key string, value string) bool {
	request := StoreRequest{key, value, root, this.addr}
	this.data.Store(request)
	request.RequesterPri = publisher
	this.RangePut(request)
	return true
}
func (this *Node) Get(key string) (bool, string) {
	isFind := false
	reply := ""
	requestInfo := FindValueRequest{key, this.addr}
	resultList := make([]Contact, 0, K*2)
	pendingList := this.table.FindClosest(Hash(key), K)
	inRun := new(int32)
	*inRun = 0
	visit := make(map[string]bool)
	visit[this.addr.Address] = true
	index := 0
	ch := make(chan FindValueReply, alpha+3)
	for index < len(pendingList) && *inRun < alpha {
		tmpReplier := pendingList[index].ContactInfo
		if _, ok := visit[tmpReplier.Address]; !ok {
			visit[tmpReplier.Address] = true
			atomic.AddInt32(inRun, 1)
			go func(target *Contact, channel chan FindValueReply) {
				var response FindValueReply
				err := RemoteCall(this, target, "WrapNode.FindValue", requestInfo, &response)
				if err != nil {
					log.Errorln("<Get> fail in Get in ", requestInfo.Request.Address)
					atomic.AddInt32(inRun, -1)
					return
				}
				channel <- response
				return
			}(&tmpReplier, ch)
		}
		index++
	}
	for (index < len(pendingList) || *inRun > 0) && !isFind {
		if *inRun > 0 {
			select {
			case response := <-ch:
				atomic.AddInt32(inRun, -1)
				if response.IsFind {
					isFind = true
					reply = response.Value
					break
				}
				resultList = append(resultList, response.Reply)
				for _, v := range response.Content {
					pendingList = append(pendingList, v)
				}
				SliceSort(&pendingList)
			case <-time.After(waitTime):
				log.Errorln("<Get> time out")
			}
			if isFind {
				break
			}
		}
		for index < len(pendingList) && *inRun < alpha {
			tmpReplier := pendingList[index].ContactInfo
			if _, ok := visit[tmpReplier.Address]; !ok {
				visit[tmpReplier.Address] = true
				atomic.AddInt32(inRun, 1)
				go func(target *Contact, channel chan FindValueReply) {
					var response FindValueReply
					err := RemoteCall(this, target, "WrapNode.FindValue", requestInfo, &response)
					if err != nil {
						log.Errorln("<Get> fail in Get in ", requestInfo.Request.Address)
						atomic.AddInt32(inRun, -1)
						return
					}
					channel <- response
					return
				}(&tmpReplier, ch)
			}
			index++
		}
	}
	if !isFind {
		return false, ""
	} else {
		StoreInfo := StoreRequest{key, reply, duplicater, this.addr}
		count := new(int32)
		*count = 0
		for i := 0; i < len(resultList); {
			if *count < alpha {
				target := resultList[i]
				i++
				atomic.AddInt32(count, 1)
				go func(targetNode *Contact, input StoreRequest) {
					var occupy string
					err := RemoteCall(this, targetNode, "WrapNode.Store", input, &occupy)
					if err != nil {
						log.Warningln("<Get> fail in Store in ", targetNode.Address)
					}
					atomic.AddInt32(count, -1)
				}(&target, StoreInfo)
			} else {
				time.Sleep(sleepTime)
			}
		}
		return true, reply
	}
}
func (this *Node) Delete(key string) bool {
	return true
}
