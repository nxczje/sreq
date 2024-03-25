package sreq

import (
	"crypto/tls"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TransportWs struct {
	conn      *websocket.Conn
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
}

func (tr *TransportWs) ReqDuration() time.Duration {
	return tr.Duration() - tr.ConnDuration()
}

func (tr *TransportWs) ConnDuration() time.Duration {
	return tr.connEnd.Sub(tr.connStart)
}

func (tr *TransportWs) Duration() time.Duration {
	return tr.reqEnd.Sub(tr.reqStart)
}

func (tr *TransportWs) Connect(urlStr string) *TransportWs {
	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	conn, _, err := dialer.Dial(urlStr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	tr.conn = conn
	return tr
}

func (tr *TransportWs) SendMessage(message string) error {
	tr.reqStart = time.Now()
	err := tr.conn.WriteMessage(websocket.TextMessage, []byte(message))
	tr.reqEnd = time.Now()
	return err
}

func (tr *TransportWs) ReadMessage() (string, time.Duration) {
	_, message, err := tr.conn.ReadMessage()
	if err != nil {
		return "", 0
	}
	return string(message), tr.Duration()
}

type Socket struct {
	Data2Send string
	DataRet   []rune
}

type indexWs struct {
	index     int
	data2Send string
	dataRet   []rune
}

type HandleWs struct {
	index     int
	Resp      string
	Data2Send string
	DataRet   []rune
	TimeRes   time.Duration
}

var CheckWs = false

func AddqueueWs(socket []Socket, tempsock Socket) []Socket {
	socket = append(socket, tempsock)
	return socket
}

// Send websocket request
func SendWS(urlStr string, queueWs []Socket, workers int, f func(handle HandleWs)) {
	var wg sync.WaitGroup
	wg.Add(workers)
	workerChanWs := make([]chan *indexWs, workers)
	for i := 0; i < workers; i++ {
		workerChanWs[i] = make(chan *indexWs, 1000)
		go func(i int) {
			defer wg.Done()
			tr := &TransportWs{}
			tr = tr.Connect(urlStr)
			defer tr.conn.Close()

			timeresp := time.Duration(0)
			resp := ""
			for indexWs := range workerChanWs[i] {
				if !CheckWs {
					tr.SendMessage(indexWs.data2Send)
					resp, timeresp = tr.ReadMessage()
					f(HandleWs{indexWs.index, resp, indexWs.data2Send, indexWs.dataRet, timeresp})
				} else {
					break
				}
			}
		}(i)
	}
	for i, v := range queueWs {
		workerChanWs[i%workers] <- &indexWs{i, v.Data2Send, v.DataRet}
	}
	for _, workerCh := range workerChanWs {
		close(workerCh)
	}
	wg.Wait()
}

// Insert new node to array and sort by Index
func AppendWs(data []HandleWs, temp HandleWs) []HandleWs {
	//Append temp to data and sort with data.Index
	index := sort.Search(len(data), func(i int) bool { return data[i].index >= temp.index })
	data = append(data, HandleWs{})
	copy(data[index+1:], data[index:])
	data[index] = temp
	return data
}
