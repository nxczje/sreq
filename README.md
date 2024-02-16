# sreq
simple request with workerpool

```
package main

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/kr/pretty"
	"github.com/nxczje/SimpReqGo/pkg/features"
	"github.com/nxczje/SimpReqGo/pkg/sreq"
)

const (
	Proxy  = "http://localhost:8080"
	Target = "http://x"
)

func NAME(){
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; rv:52.0) Gecko/20100101 Firefox/52.0")
	req.URL.Path = "/ATutor/mods/_standard/social/index_public.php"
	//if POST
	// body := pretty.Sprintf(``)
	// req, _ := http.NewRequest("POST", target, bytes.NewBuffer([]byte(body)))
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// if PUTFile
	// param := sreq.ParamInUploadFile{ParamName: "submit_import", ParamValue: "Import"}
	// bodyUpload, multipart := sreq.BodyFileUpload(sreq.FileParam{NameOfParam: "file", FileName: name}, param)
	// req, _ := http.NewRequest("POST", target, bodyUpload)
    // req4.Header.Set("Content-Type", multipart)
	// if u want segment
	// not use sreq.SPrintln. Only append
	temp := []*http.Request{}
	for i := 0; i < 100; i++ {
		req.URL.RawQuery = pretty.Sprintf(`xxxx`, i)
		temp = sreq.AddQueue(temp, &sreq.Request{Req: req, DataRet: rune(i)})
	}
	tempdata := []sreq.DataHandler{}
	sreq.Sends(temp, Proxy, false, 20, func(c sreq.DataHandler) {
		if c.Resp.StatusCode == 404 {
			tempdata = sreq.Append(tempdata, c)
		}
	})
	for _, v := range tempdata {
		pretty.Println(v.DataRet, v.Req.URL.RawQuery)
	}
}

func main() {
	NAME()
}
```