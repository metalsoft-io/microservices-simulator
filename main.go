package main

import (
	"crypto/rand"
	"flag"
	"log"
	"net/http"
	"strings"
)

//to simulate microservices, a cluster of independent microservices
//talk to eachother randomly asking for an information
//that might require a chain of microservices to retrieve
//for example: agent1 -> agent2-> agent 3
//each simulator agent will expose a single function /get_data
//that receives a chain of agents to contact to receive the data
//that the last agent exports
//agents identify each other by IP
//agents register themsevles in etcd upon their boot
//agents agent will generate multiple random payloads upon boot,
//kept in memory
//a loader application will use a monte-carlo technique to measure the impact of
//latency depending on chain length. pre-generate a list of paths to call
//will call all of them and measure the latency of each of them (sequencially)

func main() {

	interfaceName := flag.String("i", "eth0", "interface name to read IP from. A single IP must be configured.")
	ipv6 := flag.Bool("v6", false, "(flag) Use ipv6 (defaults to false)")
	etcdEndpoints := flag.String("etcd", "http://localhost:2379", "Etcd endpoints comma separated. Defaults to 127.0.0.1:2379")
	generation := flag.String("g", "0", "Used to identify between generations of microservices")
	listenOnAddrPort := flag.String("p", ":3365", "Listen ip:port. Defaults to ':8080'")
	payloadSize := flag.Int("payloadSize", 64, "Payload size in bytes (defaults to 64)")

	flag.Parse()

	//determine my ip
	myIP := getMyIP(*interfaceName, *ipv6)
	log.Printf("Discovered IP %s from interface %s", myIP, *interfaceName)

	//register myself in etcd
	registerInEtcd(myIP, *generation, strings.Split(*etcdEndpoints, ","))
	log.Printf("Registered myself in etcd at key %s/%s endpoint %s", _ETCDKey, myIP, *etcdEndpoints)

	handler := myHandler{
		myIP:    myIP,
		Payload: make([]byte, *payloadSize),
	}

	//generate random payload
	rand.Read(handler.Payload)

	log.Fatal(http.ListenAndServe(*listenOnAddrPort, &handler))
}
