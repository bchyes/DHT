package chord

import (
	"crypto/sha1"
	"math/big"
	"net"
	"time"
)

var (
	localAddress string
	base         *big.Int
	calculateMod *big.Int
	timeCut      time.Duration
	waitTime     time.Duration
)

type Pair struct {
	Key   string
	Value string
}

func init() {
	localAddress = GetLocalAddress()
	base = big.NewInt(2)
	calculateMod = new(big.Int).Exp(base, big.NewInt(160), nil)
	timeCut = 200 * time.Millisecond
	waitTime = 250 * time.Millisecond
}
func ConsistentHash(raw string) *big.Int {
	hash := sha1.New()
	hash.Write([]byte(raw))
	return (&big.Int{}).SetBytes(hash.Sum(nil))
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
func Contain(target, begin, end *big.Int, mode bool) bool {
	if end.Cmp(begin) > 0 {
		if mode {
			return end.Cmp(target) == 0 || (target.Cmp(begin) > 0 && target.Cmp(end) < 0)
		} else {
			return target.Cmp(begin) > 0 && target.Cmp(end) < 0
		}
	} else {
		if mode {
			return end.Cmp(target) == 0 || target.Cmp(begin) > 0 || target.Cmp(end) < 0
		} else {
			return target.Cmp(begin) > 0 || target.Cmp(end) < 0
		}
	}
}
func CalculateID(raw *big.Int, pow int) *big.Int {
	d := new(big.Int).Exp(base, big.NewInt(int64(pow)), calculateMod)
	ans := new(big.Int).Add(raw, d)
	return new(big.Int).Mod(ans, calculateMod)
}
