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
Like step but edit worker is 10-50.
if u send all and check result to end process u want to sort

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

If u not want to sort. Handle with segment loop

```
result :=
for xxx:
	....
	sreq.Send(){
		handle here
		result += string(x)
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

### Connect websocket with workerpool
```
	arrWs := []sreq.Socket{}
	for i := 0; i < 100; i++ {
		temp := pretty.Sprintf("42[\"newChatMessage\",{\"msg\":\"%d\"}]", i)
		arrWs = sreq.AddqueueWs(arrWs, sreq.Socket{Data2Send: temp})
	}
	sreq.SendWS(Socket, arrWs, 10, func(ws sreq.HandleWs) {
		pretty.Println(ws.TimeRes.Seconds())
	})
```

### Connect Websocket with input from command line
```
	sreq.ConnectWS(url, func(input string) string {
			sreq.ConnectWS(url, func(input string) string {
			req := //somethings format input (input is text from command line)
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