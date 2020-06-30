package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

//getPayloadFromChain contacts the first ip in the chain asking from
// payload from the last by going through all of the hops in the chain
func getPayloadFromChain(URLs []string, port int64, payloadSize int64) ([]byte, error) {

	nextHop := URLs[0]

	newChain := append([]string(nil), URLs[1:]...) //to avoid the "kind-of" memory leak

	reqDet := requestDetails{
		URLChain:    newChain,
		PayloadSize: payloadSize,
	}

	reqBody, err := json.Marshal(reqDet)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", nextHop, bytes.NewReader(reqBody))

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Could not execute HTTP request, aborting %v", err)
		return nil, err
	}

	body := &bytes.Buffer{}

	_, err = io.Copy(body, resp.Body)

	return body.Bytes(), err
}
