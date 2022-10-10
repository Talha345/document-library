package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var name = os.Getenv("NAME")
var port = "10000"
var hbase_host = "http://hbase:8080"

type TemplateData struct {
	Data       []ResultRow
	ServerName string
}

type ResultRow struct {
	Name  string
	Docs  []Document
	Metas []Meta
}

type Meta struct {
	Name  string
	Value string
}

type Document struct {
	Name  string
	Value string
}

func handleRequests(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		handleGetRequest(w, r)
	} else if r.Method == http.MethodPost {
		handlePostRequest(w, r)
	} else {
		io.WriteString(w, "This request method is not supported")
	}
}

func initializeScanner() string {
	body := strings.NewReader(`<Scanner batch="100"/>`)
	req, err := http.NewRequest("PUT", "http://hbase:8080/se2:library/scanner", body)
	if err != nil {
		log.Println("Error creating scanner", err)
	}
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("Content-Type", "text/xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error requesting scanner", err)
	}
	defer resp.Body.Close()
	return resp.Header.Get("Location")
}

func modifyResponseForHtml(rowData RowsType) []ResultRow {
	var rows []ResultRow
	for _, row := range rowData.Row {

		resRow := ResultRow{
			Name: row.Key,
		}
		for _, cell := range row.Cell {
			if strings.HasPrefix(cell.Column, "document") {
				doc := Document{
					Name:  strings.Split(cell.Column, ":")[1],
					Value: cell.Value,
				}
				resRow.Docs = append(resRow.Docs, doc)
			} else {
				meta := Meta{
					Name:  strings.Split(cell.Column, ":")[1],
					Value: cell.Value,
				}
				resRow.Metas = append(resRow.Metas, meta)
			}
		}

		rows = append(rows, resRow)
	}
	return rows
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	var scanner = initializeScanner()
	req, err := http.NewRequest("GET", scanner, nil)
	if err != nil {
		log.Println("Error GET request", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error GET request", err)
	}
	defer resp.Body.Close()
	var encodedRowData EncRowsType
	_ = json.NewDecoder(resp.Body).Decode(&encodedRowData)
	rowData, _ := encodedRowData.decode()
	htmlRow := modifyResponseForHtml(rowData)
	var tempData = TemplateData{htmlRow, name}
	templ := template.Must(template.ParseFiles("html/library.html"))
	templ.Execute(w, tempData)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {

	unencodedJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	var unencodedRows RowsType
	json.Unmarshal(unencodedJSON, &unencodedRows)
	// encode fields in Go objects
	encodedRows := unencodedRows.encode()
	// convert encoded Go objects to JSON
	encodedJSON, _ := json.Marshal(encodedRows)
	req, err := http.NewRequest("PUT", "http://hbase:8080/se2:library/fakerow", bytes.NewBuffer(encodedJSON))
	if err != nil {
		log.Println("Error POST request", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error POST request", err)
	} else {
		w.WriteHeader(200)
		log.Println("POST request successful", resp.Status)
	}
	defer resp.Body.Close()

}
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

func createNode(conn *zk.Conn) {
	var node = "/gserve/" + name
	var url = name + ":" + port
	fmt.Println(node)
	fmt.Println(url)

	namespaceExists, _, _ := conn.Exists("/gserve")
	if !namespaceExists {
		_, err := conn.Create("/gserve", []byte{}, 0, zk.WorldACL(zk.PermAll))

		if err != nil {
			log.Println("Error creating Gserve namespace", err)
		} else {
			log.Println("Gserve Namespace created")
		}
	}

	nodeExists, _, _ := conn.Exists(node)
	if !nodeExists {
		create := func() error {
			var err error
			// try creating ephemeral node
			_, err = conn.Create(node, []byte(url), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			return err
		}

		if create() != nil {
			fmt.Println("Error creating ephemeral node", create())
		} else {
			fmt.Println("Ephemeral node added" + node)
		}
	}
}

func main() {
	conn, _ := zk_connect("zookeeper")
	createNode(conn)
	http.Handle("/", http.HandlerFunc(handleRequests))
	log.Fatal(http.ListenAndServe(":10000", nil))

}
