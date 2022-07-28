package main

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/bencode-go"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var (
	pieceSize = 262144
	waitTime  = 3 * time.Second
)

type pieceInfo struct {
	index int
	data  []byte
}
type bencodeinfo struct {
	pieces      string
	piecelength int
	length      int
	name        string
}
type bencodetorrent struct {
	announce string
	info     bencodeinfo
}
type TorrentFile struct {
	announce    string
	infoHash    [20]byte
	pieceHashes [][20]byte
	pieceLength int
	length      int
	name        string
}
type downloadPackage struct {
	data  []byte
	index int
}

func (this *pieceInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *this)
	if err != nil {
		return [20]byte{}, err
	}
	//hash := sha1.New()
	//hash.Write(buf.Bytes())
	//h := hash.Sum(nil)
	h := sha1.Sum(buf.Bytes())
	return h, nil
}
func (this *bencodeinfo) infoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *this)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}
func Open(r io.Reader) (*bencodetorrent, error) {
	bto := bencodetorrent{}
	err := bencode.Unmarshal(r, &bto)
	fmt.Println(bto)
	if err != nil {
		return nil, err
	}
	return &bto, err
}
func makeTorrentFile(fileName string, targetPath string, ch3 chan string, ch4 chan bencodeinfo, ch5 chan string) {
	filestate, err := os.Stat(fileName)
	if err != nil {
		_ = fmt.Errorf(fmt.Sprintln("<MakeTorrentFile> fail to get filestate because ", err))
		return
	}
	tmp := bencodetorrent{
		announce: "DHT looks down on trackers",
		info: bencodeinfo{
			pieces:      "",
			piecelength: pieceSize,
			length:      int(filestate.Size()),
			name:        filestate.Name(),
		},
	}
	tmp.info.pieces = <-ch3
	var f *os.File
	var realFileName string
	if targetPath == "" {
		realFileName = filestate.Name() + ".torrent"
		f, _ = os.Create(realFileName)
	} else {
		realFileName = targetPath + "/" + filestate.Name() + ".torrent"
	}
	err = bencode.Marshal(f, tmp)
	fmt.Println(tmp.info.pieces)
	if err != nil {
		_ = fmt.Errorf("<MakeTorrentFile> fail in marshal")
		return
	}
	fmt.Println("<MakeTorrentFile> success generate .torrent file named ", filestate.Name()+".torrent")
	content, _ := ioutil.ReadFile(realFileName)
	ch4 <- tmp.info
	ch5 <- string(content)
	return
}
func (this *bencodeinfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(this.pieces)
	if len(buf)%hashLen != 0 {
		err := errors.New("pieceLength is illegal")
		return nil, err
	}
	numHash := len(buf) / hashLen
	hashes := make([][20]byte, numHash)
	for i := 0; i < numHash; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}
func (this *bencodetorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := this.info.infoHash()
	if err != nil {
		return TorrentFile{}, err
	}
	var pieceHashes [][20]byte
	pieceHashes, err = this.info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	t := TorrentFile{
		announce:    this.announce,
		infoHash:    infoHash,
		pieceHashes: pieceHashes,
		pieceLength: this.info.piecelength,
		length:      this.info.length,
		name:        this.info.name,
	}
	return t, err
}
