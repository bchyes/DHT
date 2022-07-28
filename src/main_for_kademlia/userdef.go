package main

import "main_for_kademlia/kademlia"

func NewNode(port int) dhtNode {
	// Todo: create a node and then return it.
	ptr := new(kademlia.Node)
	ptr.Init(port)
	return ptr
}

// Todo: implement a struct which implements the interface "dhtNode".
