package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type myHandler struct {
	myIP string
	Port int64
}

type requestDetails struct {
	URLChain    []string `json:"url_chain"`
	PayloadSize int64    `json:"payload_size"`
}

func (m *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Response = new(http.Response)

	if r.RequestURI != "/" && r.Method != "POST" {
		r.Response.StatusCode = 404
		log.Printf("%s Request to %s ignored", r.Method, r.RequestURI)
		return
	}

	body := &bytes.Buffer{}
	_, err := io.Copy(body, r.Body)
	if err != nil {
		log.Fatal(err)
	}
	//extract an ip array from the request
	var reqDet requestDetails

	err = json.Unmarshal(body.Bytes(), &reqDet)

	if err != nil {
		log.Printf("Could not parse request body :%s", string(body.Bytes()))
		r.Response = &http.Response{}
		r.Response.StatusCode = 500
		w.Write([]byte("Could not parse request body"))
		return
	}

	//if we are the last in the chain we return the payload
	//otherwise contact the next element in the chain and return the payload returned by that call
	if len(reqDet.URLChain) == 0 {

		payload := make([]byte, reqDet.PayloadSize)

		//generate random payload
		rand.Read(payload)

		w.Write(payload)
		r.Response.StatusCode = 200
	} else {

		//our next hop is the first on the list
		b, err := getPayloadFromChain(reqDet.URLChain, m.Port, reqDet.PayloadSize)

		if err != nil {
			log.Print(err)
			r.Response.StatusCode = 500
			w.Write([]byte(fmt.Sprint(err)))
			return
		}
		//..and we dump it into our reply
		w.Write(b)
	}

}
