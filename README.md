# Proxy for opening local files in the explorer/finder via link from the browser

compile via:

```sh
$ go build proxy.go
```

run as:

```sh
$ ./proxy.go -base C:/Users/Sophie/SomeFolder -token foo -port 1234
```

and put a link onto a website:

http://localhost:1234/open?name=subDir&token=foo

when the link is clicked (fetched by the browser) the subDir is open in the explorer locally on the machine where the proxy in running.

