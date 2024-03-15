# sreq
simple request with workerpool

### Only one request
```
	req, _ := http.NewRequest("POST", Target, bytes.NewBuffer([]byte(body)))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; rv:52.0) Gecko/20100101 Firefox/52.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.URL.Path = "/"
	sreq.Send(&sreq.Request{Req: req, Redirect: true}, Proxy, func(c sreq.DataHandler) {
		//handle 
	})
```

### Step by Step
Use one worker for same session

```
	temp := []*sreq.Request{}
	body := ``
	req, _ := http.NewRequest("POST", Target, bytes.NewBuffer([]byte(body)))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; rv:52.0) Gecko/20100101 Firefox/52.0")
	req.URL.Path = "/"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	temp = sreq.AddQueue(temp, &sreq.Request{Req: req, Redirect: true})
	// add request with addqueue
	// use 1 for worker
	sreq.Sends(temp, Proxy, 1, func(c sreq.DataHandler) {
		if c.Req.URL.Path == "/" {
			if c.Resp.StatusCode == 200 {
				//handle
			}
		}
	})
```

### Workerpool
Same step by step but edit worker
but result u need to sort if u using handler

```
	tempdata := []sreq.DataHandler{}
	sreq.Sends(temp, Proxy, 1, func(c sreq.DataHandler) {
		if c.Req.URL.Path == "/" {
			if c.Resp.StatusCode == 200 {
				tempdata = sreq.Append(tempdata, c)
			}
		}
	})
	for _, v := range tempdata {
		result = result + string(v.DataRet)
	}
```

### Run server tcp
Running server with time seconds and use goroutine to process before server tcp run

```
	server := server.New(&server.Config{})
	go func() {
		time.Sleep(2 * time.Second)
		sreq.Send(&sreq.Request{Req: req}, "")
	}()
	server.Run(5)
```

### Connect websocket
```
	sreq.ConnectWS(url, func(input string) string {
			sreq.ConnectWS(url, func(input string) string {
			req := //somethings
			return string(req)
		}, func(conn *websocket.Conn) {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				//worksomething
			}
		})
```

### Create request from raw request
```
req := sreq.ParseReqFile(filename,url)
```

### Trigger XSS

In feature/xss/README.md