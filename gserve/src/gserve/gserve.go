package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)
var serverName string

type rest struct {
	Host string
	Port int
}

func main() {
	conn := ZkConnection()
	defer conn.Close()
	rt := NewRest("hbase", 8080)
	http.HandleFunc("/library", rt.requestHandler)
	log.Fatal(http.ListenAndServe(":90", nil))
}

func ZkConnection() *zk.Conn{
	serverName = os.Getenv("servername")
	conn, _, err := zk.Connect([]string{"zookeeper:2181"}, 5*time.Second)
	if err != nil{
		fmt.Printf("Zookeeper Connection Error: %+v\n", err)
	}
	for{
		gServer, createErr := conn.Create("/grproxy/"+serverName, []byte(serverName+":90"), int32(1), zk.WorldACL(zk.PermAll))
		if createErr != nil{
			fmt.Printf("Zookeeper Registration error %+v\n",createErr)
		}
		if gServer != ""{
			fmt.Printf("create node: %+v\n", gServer)
			break
		}
	}

	if err != nil {
		fmt.Printf("Zookeeper Connection created error")
	}
	return conn
}

func (rt *rest) requestHandler(writer http.ResponseWriter, req *http.Request){
	if req.Method == "POST" {
		requestBody, err := ioutil.ReadAll(req.Body)
		if err !=  nil{
			fmt.Printf("Request Body Parsing Error %+v\n",err)
		}
		err = rt.Post(requestBody)
		if err != nil {
			fmt.Printf("Hbase POST error %+v\n",err)
		}
	}
	content := rt.Get()
	_,err := fmt.Fprintf(writer, "%s\n  proudly served by %s", content, serverName)
	if err != nil {
		fmt.Printf("GET Response Writer Error:  %+v\n",err)
	}
}

func NewRest(host string, port int) *rest {
	rt := new(rest)
	rt.Host = host
	rt.Port = port
	return rt
}

func (rt *rest) Get() string {
	url := "http://" + rt.Host + fmt.Sprintf(":%d/se2:library/*", rt.Port)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, getErr := client.Do(req)

	if getErr != nil {
		fmt.Printf("GET response error: %+v\n",getErr)
	}

	encodedBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("Parse error: %+v\n",err)
	}
	decodedBody := decodeJSON(encodedBody)

	return decodedBody
}


func (rt *rest) Post(requestBody []byte) error {
	encodedBody := encodeJSON(requestBody)

	url := "http://" + rt.Host +fmt.Sprintf(":%d/se2:library/fakerow", rt.Port)

	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(encodedBody)))

	if err != nil {
		fmt.Printf("POST Response Error: %+v\n", err)
		return err
	}
	fmt.Println("Post Response: ", res.Status)
	return nil
}

func encodeJSON(unencodedJSON []byte) string {
	var unencodedRows RowsType

	err := json.Unmarshal(unencodedJSON, &unencodedRows)
	if err != nil {
		fmt.Printf("Unmarshaling JSON Error %+v\n",err)
	}
	encodedRows := unencodedRows.encode()
	encodedJSON, _ := json.Marshal(encodedRows)

	return string(encodedJSON)
}

func decodeJSON(encodedJSON []byte) string {
	var encodedRows EncRowsType

	err := json.Unmarshal(encodedJSON, &encodedRows)
	if err != nil {
		fmt.Printf("Unmarshaling JSON Error %+v\n",err)
	}
	decodedRows, _ := encodedRows.decode()
	deCodedJSON, _ := json.Marshal(decodedRows)

	return string(deCodedJSON)
}