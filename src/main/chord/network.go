package chord

import (
	"errors"
	"fmt"
	//log "github.com/sirupsen///logrus"
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
			//log.Warningln("Quit Successfully")
			return
		default:
			if err != nil {
				//log.Print("rpc.Serve: accept: ", err.Error())
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

	err1 := this.serv.Register(this.nodePtr)
	if err1 != nil {
		//log.Errorf("<RPC Init> fail to register in address : %s\n", address)
		return err1
	}
	this.lis, err1 = net.Listen("tcp", address)
	if err1 != nil {
		//log.Errorf("<RPC Init> fail to listen in address : %s\n", address)
		return err1
	}
	//log.Infof("<RPC Init> service start success in %s\n", address)
	go MyAccept(this.serv, this.lis, this.nodePtr.node)
	return nil
}
func GetClient(address string, funcClass string) (*rpc.Client, error) {
	if address == "" {
		//log.Warningln("<GetClient> IP address is nil")
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
				if funcClass != "" {
					//log.Errorln("<Getclient> ", err, " in ", funcClass)
				}
				return nil, err
			}
		case <-time.After(waitTime):
			//log.Errorln("<Getclient> time out to", address)
			err = errors.New(fmt.Sprintln("<Getclient> time out to", address))
		}
	}
	return nil, err
}
func CheckOnline(address string) bool {
	client, err := GetClient(address, "")
	if err != nil {
		//log.Infoln("CheckOnline Ping Fail In", address, "beacuse ", err)
		return false
	}
	if client != nil {
		defer client.Close()
	} else {
		return false
	}
	//log.Infoln("CheckOnline Ping Success In", address)
	return true
}
func RemoteCall(targetNode string, funcClass string, input interface{}, result interface{}) error {
	if funcClass == "WrapNode.Stabilize" {
		//log.Warningln("<RemoteCall> start Stabilize1")
	}
	if targetNode == "" {
		//log.Warningln("<RemoteCall> IP address is nil")
		return errors.New("<RemoteCall> IP address is nil")
	}
	client, err := GetClient(targetNode, funcClass)
	if err != nil {
		//log.Errorln("<RemoteCall> fail to GetClient in", targetNode, "and error is", err)
		return err
	}
	if client != nil {
		defer client.Close()
	}
	if funcClass == "WrapNode.Stabilize" {
		//log.Warningln("<RemoteCall> start Stabilize2")
	}
	err1 := client.Call(funcClass, input, result)
	if funcClass == "WrapNode.Stabilize" {
		//log.Warningln("<RemoteCall> start Stabilize3")
	}
	if err1 == nil {
		//log.Infoln("<RemoteCall> in", targetNode, "with", funcClass, "success")
		return nil
	} else {
		//log.Errorln("<RemoteCall> in", targetNode, "with", funcClass, "fail because", err1)
		return err1
	}
}
func (this *network) ShutDown() error {
	this.nodePtr.node.QuitSignal <- true
	err := this.lis.Close()
	if err != nil {
		//log.Errorln("<ShutDown> Fail to close the network in ", this.nodePtr.node.address)
		return err
	}
	//log.Infoln("<ShutDown> Success in", this.nodePtr.node.address)
	return nil
}
