package main

import (
	"fmt"
	"github.com/megglymark/follicle/trnt"
	"log"
	"net/http"
	"net/url"
)

func startHandler(w http.ResponseWriter, r *http.Request) {
	m, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(m)
	query := "SELECT * FROM transfers WHERE id=" + m["id"][0]
	torrent := trnt.Torrents(query)
	t := torrent()
	for _, v := range t {
		v.TransferTorrent()
	}
}

func main() {
	http.HandleFunc("/start/", startHandler)
	http.ListenAndServe("localhost:8080", nil)
}
