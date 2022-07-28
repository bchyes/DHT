package kademlia

import (
	"crypto/sha1"
	"math/rand"
	"net"
	"sort"
	"time"
)

var (
	localAddress string
	//base         *big.Int
	//caculateMod  *big.Int
	waitTime = 250 * time.Millisecond
	//timeCut      = 200 * time.Millisecond
	sleepTime = 20 * time.Millisecond
)

func init() {
	localAddress = GetLocalAddress()
	//base = big.NewInt(2)
	//caculateMod = new(big.Int).Exp(base, big.NewInt(160), nil)
}
func NewContact(address string) Contact {
	nodeID := Hash(address)
	return Contact{address, nodeID}
}
func Hash(address string) (result ID) {
	hash := sha1.New()
	hash.Write([]byte(address))
	tmp := hash.Sum(nil)
	for i := 0; i < IDlength; i++ {
		result[i] = tmp[i]
	}
	return
}
func PreFixLen(IDvalue ID) int {
	for i := 0; i < IDlength; i++ {
		for j := 0; j < 8; j++ {
			if (IDvalue[i]>>(7-j))&(1) != 0 {
				return 8*i + j
			}
		}
	}
	return IDlength*8 - 1
}
func Xor(first ID, second ID) (result ID) {
	for i := 0; i < IDlength; i++ {
		result[i] = first[i] ^ second[i]
	}
	return
}
func (first ID) Equals(second ID) bool {
	for i := 0; i < IDlength; i++ {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}
func (first ID) LessThan(second ID) bool {
	for i := 0; i < IDlength; i++ {
		if first[i] == second[i] {
			continue
		}
		if first[i] > second[i] {
			return false
		} else {
			return true
		}
	}
	return false
}
func SliceSort(dataset *[]ContactRecord) {
	sort.Slice(*dataset, func(i int, j int) bool {
		return (*dataset)[i].SortKey.LessThan((*dataset)[j].SortKey)
	})
}
func (this *Node) GenerateID(origin ID, index int) (result ID) { ///////
	blockIndex := index / 8
	concreteIndex := index % 8
	for i := 0; i <= blockIndex; i++ {
		result[i] = origin[i]
	}
	tmp := 1
	tmp <<= (7 - concreteIndex)
	result[blockIndex] ^= byte(tmp)
	for i := blockIndex + 1; i < IDlength; i++ {
		x := rand.Uint32()
		result[i] = byte(x)
	}
	return
}
func GetLocalAddress() string {
	var localaddress string

	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}

	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}

	return localaddress
}
