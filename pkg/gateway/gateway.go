package gateway

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/Crow314/im920s-controller/pkg/module"
	"log"
	"net/http"
	"net/url"
	"time"
)

type config struct {
	address  byte
	times    int
	interval int
}

type header struct {
	Address  byte
	Service  byte
	Sequence uint16
}

type transmitRequest struct {
	service       byte
	data          []byte
	errorResponse error
	notificator   chan struct{}
}

var ServiceURL map[byte]string
var conf config
var im *module.Im920s
var transmitReqChan chan transmitRequest

var sequence uint16

func Run(im920s *module.Im920s, dataChannel <-chan module.ReceivedData, address byte, sendTimes int, interval int) {
	im = im920s
	transmitReqChan = make(chan transmitRequest)
	sequence = 0

	conf.address = address
	conf.times = sendTimes
	conf.interval = interval

	go disComReceiver(dataChannel)
	go disComTransmitter()

	http.HandleFunc("/", httpHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func disComReceiver(dataChannel <-chan module.ReceivedData) {
	for {
		data := <-dataChannel

		var h header

		err := binary.Read(bytes.NewReader(data.Data()[:4]), binary.BigEndian, &h)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}

		if h.Address == conf.address {
			continue
		}

		body := data.Data()[4:] // header: 0-3 / body: 4-

		val := url.Values{}
		val.Add("data", hex.EncodeToString(body))

		res, err := http.PostForm(ServiceURL[h.Service], val)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}

		log.Println(res.Header)
	}
}

func disComTransmitter() {
	for {
		transmitReq := <-transmitReqChan

		h := header{Address: conf.address, Service: transmitReq.service, Sequence: sequence}
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, h)
		if err != nil {
			transmitReq.errorResponse = err
			close(transmitReq.notificator) // 処理完了通知
		}

		packet := append(buf.Bytes(), transmitReq.data...)
		go transmit(packet)

		sequence++

		close(transmitReq.notificator) // 処理完了通知
	}
}

func transmit(packet []byte) {
	for i := 0; i < conf.times; i++ {
		_ = im.Broadcast(packet) // TODO error handling
		time.Sleep(time.Duration(conf.interval) * time.Millisecond)
	}
}
