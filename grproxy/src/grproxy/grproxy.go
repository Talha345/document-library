package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var nodeAddressesGlobal []string
var index int

func zk_connect(zk_server string) (*zk.Conn, error) {
	zoo_keeper, session, err := zk.Connect([]string{zk_server}, time.Second)
	if err != nil {
		panic(err)
		return nil, err
	}
	for event := range session {
		if event.State == zk.StateConnected {
			log.Printf("zookeeper State: %s\n", event.State)
			break
		}
	}
	return zoo_keeper, nil
}

func GetNodesW(conn *zk.Conn, path string) {
	nodes, _, watch, err := conn.ChildrenW(path)
	if err != nil {
		GetNodesW(conn, path)
	}
	log.Println("Nodes Changed:", len(nodes))
	var nodeAddresses []string
	for _, node := range nodes {
		if data, _, err := conn.Get(path + "/" + node); err != nil {
			log.Printf("Cannot read node data: %v", node)
		} else {
			nodeAddresses = append(nodeAddresses, string(data))
			log.Printf("node data: %v", string(data))
		}
	}
	nodeAddressesGlobal = nodeAddresses
	for {
		select {
		case event := <-watch:
			if event.Type == zk.EventNodeChildrenChanged {
				GetNodesW(conn, path)
			}

		}
	}
}

func forwardToNginx(res http.ResponseWriter, req *http.Request) {
	nginxUrl, _ := url.Parse("http://nginx:80")
	httputil.NewSingleHostReverseProxy(nginxUrl).ServeHTTP(res, req)
}

func RoundRobin(servers []string) string {
	if index >= len(servers) {
		index = 0
	}
	server := servers[index]
	index++

	return server
}

func forwardToGserve(res http.ResponseWriter, req *http.Request) {
	if len(nodeAddressesGlobal) == 0 {
		log.Println("No Server Found")
		res.WriteHeader(http.StatusServiceUnavailable)
	} else {
		serverAddr := RoundRobin(nodeAddressesGlobal)
		log.Println(serverAddr)
		serverUrl, _ := url.Parse("http://" + serverAddr)
		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = HandleErr
		proxy.ServeHTTP(res, req)
	}
}

func HandleErr(res http.ResponseWriter, req *http.Request, err error) {
	log.Println("Proxy Error:", err)
	forwardToGserve(res, req)
	//res.WriteHeader(http.StatusServiceUnavailable)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	if string(req.URL.Path) == "/library" {
		forwardToGserve(res, req)
	} else {
		forwardToNginx(res, req)
	}
}

func main() {
	conn, _ := zk_connect("zookeeper")
	go GetNodesW(conn, "/gserve")
	http.Handle("/", http.HandlerFunc(handleRequestAndRedirect))
	log.Fatal(http.ListenAndServe(":5000", nil))
}
