package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"main/chord"
	"os"
)

var (
	myself dhtNode
	myIP   string
)

func init() {
	fmt.Println("your local address is ", chord.GetLocalAddress())
	var f *os.File
	f, _ = os.Create("log.txt")
	log.SetOutput(f)
	fmt.Println("Please type your IP to quick start")
	_, _ = fmt.Scanln(&myIP)
	fmt.Println("IP is set to ", myIP, " type help to get command")
	myself = NewNode(fmt.Sprintf("%s:%s", chord.GetLocalAddress(), myIP))
	myself.Run()
	myself.Create()
}
func main() {
	for {
		var para1, para2, para3, para4 = "", "", "", ""
		_, _ = fmt.Scanln(&para1, &para2, &para3, &para4)
		if para1 == "run" {
			myself.Run()
			fmt.Println("myself run in ", myIP)
			continue
		}
		if para1 == "join" {
			ok := myself.Join(para2)
			if ok == true {
				fmt.Println("Join success through ", para2)
			} else {
				fmt.Println("Join fail through ", para2)
			}
			continue
		}
		if para1 == "create" {
			myself.Create()
			fmt.Println("Create new NetWork in ", myIP)
			continue
		}
		if para1 == "upload" {
			err := Lauch(para2, para3, &myself)
			if err != nil {
				fmt.Println("fail upload in ", myIP)
			} else {
				fmt.Println("success upload in ", myIP)
			}
			continue
		}
		if para1 == "download" {
			if para2 == "-t" {
				err := download(para3, para4, &myself)
				if err != nil {
					fmt.Println("download failed")
				}
				continue
			}
			if para2 == "-m" {
				downloadFromMagnet(para3, para4, &myself)
				continue
			}
		}
		if para1 == "quit" {
			myself.Quit()
			fmt.Println("myself quit in ", myIP)
			continue
		}
		if para1 == "help" {
			fmt.Println("<--------------------------------------------------------------------------------------------------------------------------->")
			fmt.Println("|   join [IP address]              # Join the network from IP address                                                        |")
			fmt.Println("|   run                            # Start the network before create and create                                              |")
			fmt.Println("|   create                         # Create a new node                                                                       |")
			fmt.Println("|   upload [path1] [path2]         # Upload file in path1 and generate .torrent in path2(Default is the current directory)   |")
			fmt.Println("|   download -t [path1] [path2]    # Download file by .torrent in path1 into path2(Default is current directory)             |")
			fmt.Println("|   download -m [value] [path]     # Download file by magnet into path(Default)                                              |")
			fmt.Println("|   quit                           # Quit the network                                                                        |")
			fmt.Println("<--------------------------------------------------------------------------------------------------------------------------->")
			continue
		}
		if para1 == "exit" {
			os.Exit(0)
		}
		fmt.Println("Unknown instruction")
	}
}
