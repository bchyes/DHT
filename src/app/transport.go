package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

type lauchPackage struct {
	hashes [20]byte
	index  int
}

func Lauch(fileName string, targetPath string, node *dhtNode) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Errorln("fail in read the file ", fileName)
		return err
	}
	var fileByteSize = len(content)
	var blockNum int
	if fileByteSize%pieceSize == 0 {
		blockNum = fileByteSize / pieceSize
	} else {
		blockNum = fileByteSize/pieceSize + 1
	}
	var pieces = make([]byte, 20*blockNum)
	ch1 := make(chan int, blockNum+20)
	ch2 := make(chan lauchPackage, blockNum+20)
	for i := 1; i <= blockNum; i++ {
		ch1 <- i
	}
	var flag1, flag2 = true, true
	for flag1 {
		select {
		case index := <-ch1:
			l := (index-1)*pieceSize + 1
			r := index * pieceSize
			if r > fileByteSize {
				r = fileByteSize
			}
			go uploadToNetwork(blockNum, node, index, content[l:r], ch1, ch2)
		case <-time.After(waitTime):
			flag1 = false
		}
		time.Sleep(100 * time.Millisecond)
	}
	ch3 := make(chan string)
	ch4 := make(chan bencodeinfo)
	ch5 := make(chan string)
	go makeTorrentFile(fileName, targetPath, ch3, ch4, ch5)
	for flag2 {
		select {
		case pack := <-ch2:
			index := pack.index
			copy(pieces[(index-1)*20:index*20], pack.hashes[:])
		case <-time.After(waitTime):
			fmt.Println("Upload finished, start making magnet")
			flag2 = false
		}
	}
	ch3 <- string(pieces)
	torrentInfo := <-ch4
	torrentContent := <-ch5
	fmt.Println(torrentContent)
	var temp [20]byte
	temp, err = torrentInfo.infoHash()
	infoHash := fmt.Sprintf("%x", temp)
	if err != nil {
		_ = fmt.Errorf("<Lauch> fail to get infoHash")
		return err
	}
	ok := myself.Put(infoHash, torrentContent)
	if !ok {
		_ = fmt.Errorf("<Lauch> fail to put")
		return errors.New("<Lauch> fail to put")
	}
	fmt.Println("magnet link generate successfully")
	fmt.Println("magnet:?xt=urn:sha1:" + infoHash + "&dn=" + torrentInfo.name) //?
	return nil
}
func uploadToNetwork(totalSize int, node *dhtNode, index int, data []byte, ch1 chan int, ch2 chan lauchPackage) {
	info := pieceInfo{index, data}
	hashkey, err := info.hash()
	if err != nil {
		fmt.Println("fail to hash because ", err, " try again")
		ch1 <- index
		return
	}
	var sx16 = fmt.Sprintf("%x", hashkey)
	ok := (*node).Put(sx16, string(data))
	if !ok {
		fmt.Println("fail to put try again ", index)
		ch1 <- index
		return
	}
	ch2 <- lauchPackage{hashkey, index}
	fmt.Println("Uploading ", float64(totalSize-len(ch1))/float64(totalSize)*100, "% ....")
	return
}
func download(torrentName string, targetPath string, node *dhtNode) error {
	torrentFile, err := os.Open(torrentName)
	if err != nil {
		_ = fmt.Errorf("fail to open torrnetFile when downloading")
		return err
	}
	fmt.Println("open success")
	fmt.Println(torrentFile)
	var content1 []byte
	content1, _ = ioutil.ReadFile(torrentName)
	fmt.Println(string(content1))
	var BT *bencodetorrent
	BT, err = Open(torrentFile)
	if err != nil {
		_ = fmt.Errorf("fail to unmarshal the .torrent file")
		return err
	}
	fmt.Println("torrentFile unmarshal")
	fmt.Println(BT.info.pieces)
	var allInfo TorrentFile
	allInfo, err = BT.toTorrentFile()
	if err != nil {
		_ = fmt.Errorf("fail to change to .torrent file")
		return err
	}
	fmt.Println("start downloading ", allInfo.name)
	var content = make([]byte, allInfo.length)
	var blocksize = len(allInfo.pieceHashes)
	fmt.Println(blocksize)
	ch1 := make(chan int, blocksize+5)
	ch2 := make(chan downloadPackage, blocksize+5)
	for i := 1; i <= blocksize; i++ {
		ch1 <- i
	}
	flag1, flag2 := true, true
	for flag1 {
		select {
		case index := <-ch1:
			go downloadFromNetwork(blocksize, node, allInfo.pieceHashes[index-1], index, ch1, ch2)
		case <-time.After(waitTime):
			fmt.Println("download finished")
			flag1 = false
		}
	}
	for flag2 {
		select {
		case pack := <-ch2:
			l := allInfo.pieceLength * (pack.index - 1)
			r := allInfo.pieceLength * pack.index
			if r > allInfo.length {
				r = allInfo.length
			}
			copy(content[l:r], pack.data[:])
		case <-time.After(waitTime):
			var dfile string
			if targetPath == "" {
				dfile = allInfo.name
			} else {
				dfile = targetPath + "/" + allInfo.name
			}
			err := ioutil.WriteFile(dfile, content, 0644)
			if err == nil {
				fmt.Println("fail to write to file")
				return err
			}
			flag2 = false
		}
	}
	return nil
}
func downloadFromNetwork(totalSize int, node *dhtNode, hashSet [20]byte, index int, ch1 chan int, ch2 chan downloadPackage) {
	var sx16 = fmt.Sprintf("%x", hashSet)
	ok, rawData := (*node).Get(sx16)
	if !ok {
		fmt.Println("fail to download, try again")
		ch1 <- index
		return
	}
	info := pieceInfo{index, []byte(rawData)}
	verifyHash, err := info.hash()
	if err != nil {
		fmt.Println("fail to hash when verify")
		ch1 <- index
		return
	}
	if verifyHash != hashSet {
		fmt.Println("fail to verify, try again")
		ch1 <- index
		return
	}
	ch2 <- downloadPackage{[]byte(rawData), index}
	fmt.Println("Downloading ", float64(totalSize-len(ch1))/float64(totalSize)*100, "% ....")
	return
}
func downloadFromMagnet(rawMagnet string, targetPath string, node *dhtNode) {
	raw := []byte(rawMagnet)
	// sha1: 40 hex number
	//magnet:?xt=urn:sha1:    [0:20]
	infoHash := string(raw[20:60])
	ok, content := (*node).Get(infoHash)
	if !ok {
		fmt.Println("fail to get .torrent file by magnet")
		return
	}
	err := ioutil.WriteFile("temp.torrent", []byte(content), 0644)
	defer os.Remove("temp.torrent")
	if err != nil {
		fmt.Println("fail to create temporary file")
		return
	}
	err = download("temp.torrent", targetPath, node)
	if err != nil {
		fmt.Println("fail to download")
		return
	}
	return
}
