package kademlia

import (
	"container/list"
	"sync"
	"time"
)

type RoutingTable struct {
	nodeAddr       Contact
	rwLock         sync.RWMutex
	bucket         [IDlength * 8]*list.List
	refreshIndex   int
	refreshTimeSet [IDlength * 8]time.Time
}

func (this *RoutingTable) InitRoutingTable(addr Contact) {
	this.nodeAddr = addr
	this.rwLock.Lock()
	for i := 0; i < IDlength*8; i++ {
		this.bucket[i] = list.New()
		this.refreshTimeSet[i] = time.Now()
	}
	this.refreshIndex = 0
	this.rwLock.Unlock()
}
func (this *RoutingTable) Update(addr *Contact) {
	this.rwLock.RLock()
	bucket := this.bucket[PreFixLen(Xor(this.nodeAddr.NodeID, addr.NodeID))]
	target := bucket.Front() //////////
	target = nil
	for i := bucket.Front(); ; i = i.Next() {
		if i == nil {
			target = nil
			break
		}
		if i.Value.(*Contact).NodeID.Equals(addr.NodeID) {
			target = i
			break
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	if target != nil {
		bucket.MoveToBack(target)
	} else {
		if bucket.Len() < K {
			bucket.PushBack(addr)
		} else {
			tmp := bucket.Front()
			if !PureCheckConn(tmp.Value.(*Contact).Address) {
				bucket.Remove(tmp)
				bucket.PushBack(addr)
			} else {
				bucket.MoveToBack(tmp)
			}
		}
	}
	this.rwLock.Unlock()
}
func (this *RoutingTable) FindClosest(targetID ID, count int) []ContactRecord {
	result := make([]ContactRecord, 0, count)
	index := PreFixLen(Xor(this.nodeAddr.NodeID, targetID))
	this.rwLock.RLock()
	if targetID == this.nodeAddr.NodeID {
		result = append(result, ContactRecord{Xor(targetID, this.nodeAddr.NodeID), NewContact(this.nodeAddr.Address)})
	}
	for i := this.bucket[index].Front(); i != nil && len(result) < count; i = i.Next() {
		if !PureCheckConn(i.Value.(*Contact).Address) {
			continue
		}
		contact := i.Value.(*Contact)
		result = append(result, ContactRecord{Xor(targetID, contact.NodeID), *contact})
	}
	for i := 1; (index+i < IDlength*8 || index-i >= 0) && len(result) < count; i++ {
		if index+i < IDlength*8 {
			for j := this.bucket[index+i].Front(); j != nil && len(result) < count; j = j.Next() {
				if !PureCheckConn(j.Value.(*Contact).Address) {
					continue
				}
				contact := j.Value.(*Contact)
				result = append(result, ContactRecord{Xor(targetID, contact.NodeID), *contact})
			}
		}
		if index-i >= 0 {
			for j := this.bucket[index-i].Front(); j != nil && len(result) < count; j = j.Next() {
				if !PureCheckConn(j.Value.(*Contact).Address) {
					continue
				}
				contact := j.Value.(*Contact)
				result = append(result, ContactRecord{Xor(targetID, contact.NodeID), *contact})
			}
		}
	}
	this.rwLock.RUnlock()
	return result
}
