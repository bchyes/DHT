package kademlia

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	root                  = 0
	publisher             = 1
	duplicater            = 2
	common                = 3
	republicTimeInterval  = 5 * time.Hour
	duplicateTimeInterval = 15 * time.Minute
	expireTimeInterval2   = 6 * time.Hour
	expireTimeInterval3   = 20 * time.Minute
)

type database struct {
	rwLock        sync.RWMutex
	dataset       map[string]string
	expireTime    map[string]time.Time
	duplicateTime map[string]time.Time
	republicTime  map[string]time.Time
	privilege     map[string]int
}

func (this *database) Init() {
	this.rwLock.Lock()
	this.dataset = make(map[string]string)
	this.expireTime = make(map[string]time.Time)
	this.duplicateTime = make(map[string]time.Time)
	this.republicTime = make(map[string]time.Time)
	this.privilege = make(map[string]int)
	this.rwLock.Unlock()
}
func (this *database) Store(request StoreRequest) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	if _, ok := this.dataset[request.Key]; !ok {
		requestPri := request.RequesterPri
		this.privilege[request.Key] = requestPri + 1
		this.dataset[request.Key] = request.Value
		if requestPri == root {
			this.republicTime[request.Key] = time.Now().Add(republicTimeInterval)
			return
		} else if requestPri == publisher {
			this.duplicateTime[request.Key] = time.Now().Add(duplicateTimeInterval)
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			return
		} else if requestPri == duplicater {
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			return
		}
		log.Errorln("<Store> fail in store ", request.Key)
	} else {
		originPri := this.privilege[request.Key]
		requestPri := request.RequesterPri
		if requestPri+1 >= originPri {
			return
		}
		this.privilege[request.Key] = requestPri + 1
		this.dataset[request.Key] = request.Value
		if requestPri == root {
			this.republicTime[request.Key] = time.Now().Add(republicTimeInterval)
			delete(this.duplicateTime, request.Key)
			delete(this.expireTime, request.Key)
			return
		} else if requestPri == publisher {
			this.duplicateTime[request.Key] = time.Now().Add(duplicateTimeInterval)
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval2)
			delete(this.republicTime, request.Key)
			return
		} else if requestPri == duplicater {
			this.expireTime[request.Key] = time.Now().Add(expireTimeInterval3)
			delete(this.duplicateTime, request.Key)
			delete(this.republicTime, request.Key)
			return
		}
		log.Errorln("<Store> fail in store ", request.Key)
	}
}
func (this *database) republic() (result map[string]string) {
	result = make(map[string]string)
	this.rwLock.RLock()
	for k, v := range this.republicTime {
		if !v.After(time.Now()) {
			result[k] = this.dataset[k]
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range result {
		this.republicTime[k] = time.Now().Add(republicTimeInterval)
	}
	this.rwLock.Unlock()
	return
}
func (this *database) duplicate() (result map[string]string) {
	result = make(map[string]string)
	this.rwLock.RLock()
	for k, v := range this.duplicateTime {
		if !v.After(time.Now()) {
			result[k] = this.dataset[k]
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range result {
		this.duplicateTime[k] = time.Now().Add(duplicateTimeInterval)
	}
	this.rwLock.Unlock()
	return
}
func (this *database) expire() {
	tmp := make(map[string]bool)
	this.rwLock.RLock()
	for k, v := range this.expireTime {
		if !v.After(time.Now()) {
			tmp[k] = true
		}
	}
	this.rwLock.RUnlock()
	this.rwLock.Lock()
	for k, _ := range tmp {
		delete(this.dataset, k)
		delete(this.republicTime, k)
		delete(this.duplicateTime, k)
		delete(this.expireTime, k)
		delete(this.privilege, k)
	}
	this.rwLock.Unlock()
}
func (this *database) get(key string) (bool, string) {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	if v, ok := this.dataset[key]; ok {
		if _, ok1 := this.expireTime[key]; ok1 {
			if !this.expireTime[key].After(time.Now().Add(expireTimeInterval3)) {
				this.expireTime[key] = time.Now().Add(expireTimeInterval3)
			}
		}
		return true, v
	}
	return false, ""
}
