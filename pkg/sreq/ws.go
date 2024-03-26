package sreq

import (
	"bufio"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
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

func (tr *TransportWs) Connect(urlStr string, proxy string) *TransportWs {
	dialer := &websocket.Dialer{}
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		dialer = &websocket.Dialer{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyURL),
		}
	} else {
		dialer = &websocket.Dialer{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
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

func (tr *TransportWs) ReadMessage(messageChan chan respChan, timeConn int) {
	if timeConn != 0 {
		tr.conn.SetReadDeadline(time.Now().Add(time.Duration(timeConn) * time.Second))
	} else {
		//set 1s timeout
		tr.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	}

	for {
		_, message, err := tr.conn.ReadMessage()
		if err != nil {
			break
		}
		messageChan <- respChan{Resp: string(message), TimeRes: tr.Duration()}
	}
	close(messageChan)
}

type Socket struct {
	Data2Send string
	DataRet   []rune //Default nil
	TimeConn  int    //Default 1s
}

type indexWs struct {
	index     int
	data2Send string
	dataRet   []rune
	timeconn  int
}

type HandleWs struct {
	index     int
	Resp      string
	Data2Send string
	DataRet   []rune
	TimeRes   time.Duration
}

type respChan struct {
	Resp    string
	TimeRes time.Duration
}

var CheckWs = false //Check if all request is done -> stop process send request

// Add queue to send websocket request
func AddqueueWs(socket []Socket, tempsock Socket) []Socket {
	socket = append(socket, tempsock)
	return socket
}

// Send websocket request with workerpool
func SendWS(urlStr string, queueWs []Socket, workers int, proxy string, f func(handle HandleWs)) {
	CheckWs = false
	var wg sync.WaitGroup
	wg.Add(workers)
	workerChanWs := make([]chan *indexWs, workers)
	for i := 0; i < workers; i++ {
		workerChanWs[i] = make(chan *indexWs, 1000)
		go func(i int) {
			defer wg.Done()
			for indexWs := range workerChanWs[i] {
				if !CheckWs {
					tr := &TransportWs{}
					tr = tr.Connect(urlStr, proxy)
					defer tr.conn.Close()
					err := tr.SendMessage(indexWs.data2Send)
					if err != nil {
						log.Println(err)
					}
					messages := make(chan respChan)
					go tr.ReadMessage(messages, indexWs.timeconn)
					for message := range messages {
						f(HandleWs{indexWs.index, message.Resp, indexWs.data2Send, indexWs.dataRet, message.TimeRes})
					}
				} else {
					break
				}
			}
		}(i)
	}
	for i, v := range queueWs {
		workerChanWs[i%workers] <- &indexWs{i, v.Data2Send, v.DataRet, v.TimeConn}
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

// Use this for input from command line to send data
func ConnectWS(urlStr string, ft func(input string) string, f func(conn *websocket.Conn)) {
	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	conn, _, err := dialer.Dial(urlStr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()
	go f(conn)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		data := ft(text)
		err := conn.WriteMessage(websocket.TextMessage, []byte(data))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}
