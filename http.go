package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPServer() chan string {
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	sdpChan := make(chan string)
	http.HandleFunc("/post/sdp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, _ := ioutil.ReadAll(r.Body)
			sdpChan <- string(body)
			answer := <- sdpChan
			w.Write([]byte(answer))
			fmt.Println("use fmt")
			//r.ParseForm()
			//w.WriteHeader(http.StatusOK)
			//w.Header().Add("Connection", "close")
			//fmt.Fprintf(w, "{}\n", answer)
			//r.Body.Close()
		} else {
			http.Error(w, "Invalid Method", 405)
		}
	})

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
		if err != nil {
			panic(err)
		}
	}()

	return sdpChan
}
