package kademlia

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"time"
)

type network struct {
	serv    *rpc.Server
	lis     net.Listener
	nodePtr *WrapNode
}

func MyAccept(server *rpc.Server, lis net.Listener, ptr *Node) {
	for {
		conn, err := lis.Accept()
		select {
		case <-ptr.QuitSignal:
			return
		default:
			if err != nil {
				log.Errorln("<MyAccept> fail in accept because ", err)
				return
			}
			go server.ServeConn(conn)
		}
	}
}
func (this *network) Init(address string, ptr *Node) error {
	this.serv = rpc.NewServer()
	this.nodePtr = new(WrapNode)
	this.nodePtr.node = ptr
	err := this.serv.Register(this.nodePtr)
	//log.Warningln("<init> in ", address)
	if err != nil {
		log.Errorln("<Init> can't register in ", address)
		return err
	}
	this.lis, err = net.Listen("tcp", address)
	if err != nil {
		log.Errorln("<Init> can't listen in ", address)
		return err
	}
	go MyAccept(this.serv, this.lis, this.nodePtr.node)
	return nil
}
func GetClient(address string) (*rpc.Client, error) {
	if address == "" {
		log.Errorln("<GetClient> fail because address is nil")
		return nil, errors.New("<GetClient> fail because address is nil")
	}
	var client *rpc.Client
	var err error
	ch := make(chan error)
	for i := 0; i < 5; i++ {
		go func() {
			client, err = rpc.Dial("tcp", address)
			//if err != nil {
			//	println(err)
			//}
			ch <- err
		}()
		select {
		case <-ch:
			if err != nil {
				//println(err)
				return nil, err
			} else {
				return client, nil
			}
		case <-time.After(waitTime):
			err = errors.New(fmt.Sprintln("<GetClient> timeOut in GetClient in ", address))
		}
	}
	return nil, err
}
func CheckOnline(self *Node, target *Contact) bool {
	var occupy string
	err := RemoteCall(self, target, "WrapNode.Ping", self.addr, &occupy)
	if err != nil {
		log.Errorln("<CheckOnline> fail in Ping in ", target.Address)
		return false
	}
	return true
}
func PureCheckConn(target string) bool {
	client, err := rpc.Dial("tcp", target)
	if err != nil {
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}
	return true
}
func RemoteCall(self *Node, target *Contact, funcClass string, in interface{}, result interface{}) error {
	if target.Address == "" {
		return errors.New("<RemoteCall> tcp address is nil")
	}
	client, err := GetClient(target.Address)
	if err != nil {
		return nil
	}
	if client != nil {
		self.table.Update(target) /////
		defer client.Close()
	}
	err = client.Call(funcClass, in, result)
	if err != nil {
		return nil
	} else {
		return err
	}

}
func (this *network) ShutDown() error {
	this.nodePtr.node.QuitSignal <- true
	//log.Warningln("ShutDown1")
	time.Sleep(3 * sleepTime)
	err := this.lis.Close()
	//log.Warningln("ShutDown2")
	if err != nil {
		return err
	} else {
		return nil
	}
}
