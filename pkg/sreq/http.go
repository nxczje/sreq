package sreq

import (
	"bufio"
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

func spawnRequest(request *http.Request, proxy string, redirect bool, jar *cookiejar.Jar) (*http.Response, time.Duration) {
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

// Process Upload File with more param
type ParamInUploadFile struct {
	ParamName  string
	ParamValue string
}

// Process Upload File with file
type FileParam struct {
	NameOfParam string //param of filename
	FileName    string //filename to open
}

// Process Upload file body in Post request
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

func fileUpload(request *http.Request, multipath string, proxy string, redirect bool, jar *cookiejar.Jar) (*http.Response, time.Duration) {
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

// Data of response store
type DataHandler struct {
	index   int
	Resp    *http.Response
	Req     *http.Request
	Time    time.Duration
	DataRet []rune
}

type indexOfWorker struct {
	index    int
	req      *http.Request
	redirect bool
	dataret  []rune
}

// Create Request to worker
type Request struct {
	Req      *http.Request
	Redirect bool   // default false
	DataRet  []rune // default nil
}

// Check = True to stop workerpool
var Check bool

// One worker send request
func Send(req *Request, proxy string) DataHandler {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	resp := &http.Response{}
	timeresp := time.Duration(0)
	if strings.Contains(req.Req.Header.Get("Content-Type"), "multipart/form-data") {
		resp, timeresp = fileUpload(req.Req, req.Req.Header.Get("Content-Type"), proxy, req.Redirect, jar)
	} else {
		resp, timeresp = spawnRequest(req.Req, proxy, req.Redirect, jar)
	}
	return DataHandler{0, resp, req.Req, timeresp, nil}
}

// Multi worker to send request
func Sends(req []*Request, proxy string, worker int, f func(c DataHandler)) {
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
						resp, timeresp = fileUpload(r.req, r.req.Header.Get("Content-Type"), proxy, r.redirect, jar)
					} else {
						resp, timeresp = spawnRequest(r.req, proxy, r.redirect, jar)
					}
					f(DataHandler{r.index, resp, r.req, timeresp, r.dataret})
				} else {
					break
				}
			}
		}(i)
	}
	for i, v := range req {
		workerChannels[i%worker] <- &indexOfWorker{i, v.Req, v.Redirect, v.DataRet}
	}
	for _, workerCh := range workerChannels {
		close(workerCh)
	}
	wg.Wait()
}

// Add queue to run workerpool
func AddQueue(req []*Request, tempreq *Request) []*Request {
	req = append(req, &Request{tempreq.Req.Clone(tempreq.Req.Context()), tempreq.Redirect, tempreq.DataRet})
	return req
}

// Insert new node to array and sort by Index
func Append(data []DataHandler, temp DataHandler) []DataHandler {
	//Append temp to data and sort with data.Index
	index := sort.Search(len(data), func(i int) bool { return data[i].index >= temp.index })
	data = append(data, DataHandler{})
	copy(data[index+1:], data[index:])
	data[index] = temp
	return data
}

// Set cookie for request
func SetCookie(req *http.Request, cookies []*http.Cookie) *http.Request {
	if len(cookies) == 0 {
		return req
	} else {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		return req
	}
}

// Parse request from file raw
//
// url have http:// or https://
func ParseReqFile(filename string, url string) *http.Request {
	var header_collection_done bool
	headerMap := make(map[string]string)
	Post_data := ""
	var (
		method string
		path   string
		query  string
	)
	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("Open file error: %v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if !header_collection_done {
			if sc.Text() == "" {
				header_collection_done = true
			} else {
				substrings := strings.Split(sc.Text(), ":")
				if len(substrings) < 2 {
					temp := strings.Split(sc.Text(), " ")
					method = temp[0]
					if strings.Contains(temp[1], "?") {
						temp_query := strings.Split(temp[1], "?")
						path = temp_query[0]
						query = temp_query[1]
					} else {
						path = temp[1]
					}
				} else if len(substrings) > 2 {
					headerMap[substrings[0]] = strings.Trim(strings.Join(substrings[1:], ":"), " ")
				} else {
					headerMap[substrings[0]] = strings.Trim(substrings[1], " ")
				}
			}
		} else {
			Post_data = Post_data + sc.Text()
		}
	}
	f.Close()
	if err := sc.Err(); err != nil {
		log.Fatalf("Scan file error: %v", err)
	}
	// req, err := http.NewRequest(method, path+"?"+query, strings.NewReader(Post_data))
	req, err := http.NewRequest(method, url, strings.NewReader(Post_data))
	req.URL.Path = path
	req.URL.RawQuery = query
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range headerMap {
		req.Header.Set(key, value)
	}
	return req
}
