package sreq

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kr/pretty"
)

type CustomTransport struct {
	rtp       http.RoundTripper
	dialer    *net.Dialer
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
}

// Handle Time in request
func (tr *CustomTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.reqStart = time.Now()
	resp, err := tr.rtp.RoundTrip(r)
	tr.reqEnd = time.Now()
	return resp, err
}

func (tr *CustomTransport) Dial(network, addr string) (net.Conn, error) {
	tr.connStart = time.Now()
	cn, err := tr.dialer.Dial(network, addr)
	tr.connEnd = time.Now()
	return cn, err
}

func (tr *CustomTransport) ReqDuration() time.Duration {
	return tr.Duration() - tr.ConnDuration()
}

func (tr *CustomTransport) ConnDuration() time.Duration {
	return tr.connEnd.Sub(tr.connStart)
}

func (tr *CustomTransport) Duration() time.Duration {
	return tr.reqEnd.Sub(tr.reqStart)
}

// NewTransport return CustomTransport
func NewTransport(proxy string) *CustomTransport {
	tr := &CustomTransport{
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: time.Second,
			DualStack: true,
		},
	}
	if proxy == "" {
		tr.rtp = &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			DisableCompression: true,
			Dial:               tr.Dial,
		}
	} else {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		proxy := http.ProxyURL(proxyURL)
		tr.rtp = &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			DisableCompression: true,
			Dial:               tr.Dial,
			Proxy:              proxy,
		}
	}
	return tr
}

func SpawnRequest(request *http.Request, proxy string, redirect bool, jar *cookiejar.Jar) (*http.Response, time.Duration) {
	tr := NewTransport(proxy)
	client := &http.Client{}
	if jar == nil {
		if redirect {
			client = &http.Client{
				Transport: tr,
			}
		} else {
			client = &http.Client{
				Transport: tr,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
		}
	} else {
		if redirect {
			client = &http.Client{
				Transport: tr,
				Jar:       jar,
			}
		} else {
			client = &http.Client{
				Transport: tr,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
				Jar: jar,
			}
		}
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return resp, tr.Duration()
}

// Process Upload File
type ParamInUploadFile struct {
	ParamName  string
	ParamValue string
}

type FileParam struct {
	NameOfParam string
	FileName    string
}

func BodyFileUpload(fileUpload FileParam, Params ...ParamInUploadFile) (*bytes.Buffer, string) {
	file, err := os.Open(fileUpload.FileName)
	if err != nil {
		pretty.Println("Error opening file:", err)
		return nil, ""
	}
	defer file.Close()
	body := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(body)
	fileWriter, err := multipartWriter.CreateFormFile(fileUpload.NameOfParam, fileUpload.FileName)
	if err != nil {
		pretty.Println("Error creating form file:", err)
		return nil, ""
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		pretty.Println("Error copying file data:", err)
		return nil, ""
	}
	for _, param := range Params {
		multipartWriter.WriteField(param.ParamName, param.ParamValue)
	}
	multipartWriter.Close()
	return body, multipartWriter.FormDataContentType()
}

func FileUpload(request *http.Request, multipath string, proxy string, redirect bool, jar *cookiejar.Jar) (*http.Response, time.Duration) {
	tr := NewTransport(proxy)
	client := &http.Client{}
	if jar == nil {
		if redirect {
			client = &http.Client{
				Transport: tr,
			}
		} else {
			client = &http.Client{
				Transport: tr,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
		}
	} else {
		if redirect {
			client = &http.Client{
				Transport: tr,
				Jar:       jar,
			}
		} else {
			client = &http.Client{
				Transport: tr,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
				Jar: jar,
			}
		}
	}
	request.Header.Set("Content-Type", multipath)
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return resp, tr.Duration()
}

// read body response
func UnzipResp(resp *http.Response) string {
	var reader io.ReadCloser
	var err error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
	case "deflate":
		reader = flate.NewReader(resp.Body)
	default:
		reader = resp.Body
	}

	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyBytes)
}

// Process with worker

type DataHandler struct {
	Index   int
	Resp    *http.Response
	Req     *http.Request
	Time    time.Duration
	DataRet rune
}

type indexOfWorker struct {
	index   int
	req     *http.Request
	dataret rune
}

type Request struct {
	Req     *http.Request
	DataRet rune
}

var Check bool

func Sends(req []*Request, proxy string, redirect bool, worker int, f func(c DataHandler)) {
	Check = false
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	workerChannels := make([]chan *indexOfWorker, worker)
	var wg sync.WaitGroup
	for i := 0; i < worker; i++ {
		wg.Add(1)
		workerChannels[i] = make(chan *indexOfWorker, 1000)
		go func(i int) {
			defer wg.Done()
			resp := &http.Response{}
			timeresp := time.Duration(0)
			for r := range workerChannels[i] {
				if !Check {
					if strings.Contains(r.req.Header.Get("Content-Type"), "multipart/form-data") {
						resp, timeresp = FileUpload(r.req, r.req.Header.Get("Content-Type"), proxy, redirect, jar)
					} else {
						resp, timeresp = SpawnRequest(r.req, proxy, redirect, jar)
					}
					f(DataHandler{r.index, resp, r.req, timeresp, r.dataret})
				} else {
					break
				}
			}
		}(i)
	}
	for i, v := range req {
		workerChannels[i%worker] <- &indexOfWorker{i, v.Req, v.DataRet}
	}
	for _, workerCh := range workerChannels {
		close(workerCh)
	}
	wg.Wait()
}

func AddQueue(req []*Request, tempreq *Request) []*Request {
	req = append(req, &Request{tempreq.Req.Clone(tempreq.Req.Context()), tempreq.DataRet})
	return req
}

// Insert new node to array and sort by Index
func Append(data []DataHandler, temp DataHandler) []DataHandler {
	//Append temp to data and sort with data.Index
	index := sort.Search(len(data), func(i int) bool { return data[i].Index >= temp.Index })
	data = append(data, DataHandler{})
	copy(data[index+1:], data[index:])
	data[index] = temp
	return data
}
