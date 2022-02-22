package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

var lastVisitedServer = 0

func NewMultipleHostsReverseProxy(conn *zk.Conn) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		fmt.Printf("Received request %s %s %s \n", req.Method, req.Host, req.RemoteAddr)
		if req.URL.Path == "/library" {
			if req.Method == "POST" {
				reDeliverRequest(conn,req,"application/json")
			}else if req.Method == "GET"{
				reDeliverRequest(conn,req,"text/html")
			}
		}else{
			req.URL.Scheme = "http"
			req.URL.Host = "nginx"
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func reDeliverRequest(conn *zk.Conn,req *http.Request, contentType string){
	serverLists, err := GetServerLists(conn, "/grproxy")
	fmt.Printf("%#v\n", serverLists)
	if err != nil {
		fmt.Printf("zookeeper acquire server list error:%+v \n", err)
	}
	var hosts []string
	for _, server := range serverLists {
		gServerHost, _, err := conn.Get("/grproxy/" + server)
		hosts = append(hosts, string(gServerHost))
		if err != nil {
			fmt.Printf("Zookeeper get host error: %+v\n", err)
		}
	}
	if len(hosts) > 0 {
		lastVisitedServer += 1
		target := hosts[lastVisitedServer%len(hosts)]
		fmt.Printf("Target Server: %s\n", target)
		lastVisitedServer = lastVisitedServer % len(hosts)
		req.URL.Scheme = "http"
		req.URL.Host = target
		req.Header.Set("Content-Type", contentType)
	}
}

func ZKStartup()(conn *zk.Conn){
	conn, _, err := zk.Connect([]string{"zookeeper:2181"}, 5*time.Second)

	if err != nil {
		fmt.Printf("Zookeeper Connection creation error %+v\n",err)
	}
	RegisterGrproxy(conn)
	return conn
}


func RegisterGrproxy(conn *zk.Conn){
	fmt.Printf(" gproxy regisration starts \n")

	grproxy, err := conn.Create("/grproxy", []byte("grproxy:80"), int32(0), zk.WorldACL(zk.PermAll))

	if err !=nil {
		fmt.Printf("grproxy root created error: %+v\n", err)
	}
	fmt.Printf("/grproxy create: %+v\n", grproxy)
	fmt.Printf("grproxy registration completed\n")
}

func GetServerLists(conn *zk.Conn,path string) (list []string, err error) {
	list, _, err = conn.Children(path)
	return list,err
}

func main() {
	fmt.Println("Nginx Serve on :80")
	conn := ZKStartup()
	defer conn.Close()
	ReverseProxy := NewMultipleHostsReverseProxy(conn)
	log.Fatal(http.ListenAndServe(":80", ReverseProxy))
}