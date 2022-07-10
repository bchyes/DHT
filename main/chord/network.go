package chord

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
	nodePtr *wrapNode
}

func MyAccept(server *rpc.Server, lis net.Listener, ptr *Node) {
	for {
		conn, err := lis.Accept()
		select {
		case <-ptr.QuitSignal:
			return
		default:
			if err != nil {
				log.Print("rpc.Serve: accept: ", err.Error())
				return
			}
			go server.ServeConn(conn)
		}
	}
}
func (this *network) Init(address string, ptr *Node) error {
	this.serv = rpc.NewServer()
	this.nodePtr = new(wrapNode)
	this.nodePtr.node = ptr

	err1 := this.serv.Register(this.nodePtr)
	if err1 != nil {
		log.Errorf("<RPC Init> fail to register in address : %s\n", address)
		return err1
	}
	this.lis, err1 = net.Listen("tcp", address)
	if err1 != nil {
		log.Errorf("<RPC Init> fail to listen in address : %s\n", address)
		return err1
	}
	log.Infof("<RPC Init> service start success in %s\n", address)
	go MyAccept(this.serv, this.lis, this.nodePtr.node)
	return nil
}
func GetClient(address string) (*rpc.Client, error) {
	if address == "" {
		log.Warningln("<GetClient> IP address is nil")
		return nil, errors.New("<GetClient> IP address is nil")
	}
	var client *rpc.Client
	var err error
	ch := make(chan error)
	for i := 0; i < 5; i++ {
		go func() {
			client, err = rpc.Dial("tcp", address)
			ch <- err
		}()
		select {
		case <-ch:
			if err == nil {
				return client, nil
			} else {
				log.Errorln("<Getclient> ", err)
				return nil, err
			}
		case <-time.After(waitTime):
			log.Errorln("<Getclient> time out to", address)
			err = errors.New(fmt.Sprintln("<Getclient> time out to", address))
		}
	}
	return nil, err
}
func RemoteCall(targetNode string, funcClass string, input interface{}, result interface{}) error {
	if targetNode == "" {
		log.Warningln("<RemoteCall> IP address is nil")
		return errors.New("<RemoteCall> IP address is nil")
	}
	client, err := GetClient(targetNode)
	if err != nil {
		log.Errorln("<RemoteCall> fail to GetClient in", targetNode, "and error is", err)
	}
	if client != nil {
		defer client.Close()
	}
	err1 := client.Call(funcClass, input, result)
	if err1 == nil {
		log.Infoln("<RemoteCall> in", targetNode, "with", funcClass, "success")
		return nil
	} else {
		log.Infoln("<RemoteCall> in", targetNode, "with", funcClass, "fail because", err1)
		return err1
	}
}
