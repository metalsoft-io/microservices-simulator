package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gonum.org/v1/gonum/stat/combin"
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

func serverCmd(interfaceName string, ipv6 bool, etcdEndpoints []string, generation string, listenOnIP string, port int64, payloadSize int64) error {
	//determine my ip
	myIP := getMyIP(interfaceName, ipv6)
	log.Printf("Discovered IP %s from interface %s", myIP, interfaceName)

	finished := make(chan bool)
	log.Printf("Registering %s in etcd %v", myIP, etcdEndpoints)

	//register myself in etcd and start the renew lease loop
	go registerInEtcdAndRenewLeases(myIP, port, etcdEndpoints, finished)

	//start http server
	go startHTTPServer(listenOnIP, port, myIP, payloadSize)

	//forever loop of the main thread
	for {
		time.Sleep(5 * time.Second)
	}

}

func startHTTPServer(listenOnIP string, port int64, myIP string, payloadSize int64) {

	handler := myHandler{
		myIP:    myIP,
		Payload: make([]byte, payloadSize),
	}

	handler.Port = port
	//generate random payload
	rand.Read(handler.Payload)

	listenOnAddrPort := fmt.Sprintf("%s:%d", listenOnIP, port)
	log.Printf("Listening on %s", listenOnAddrPort)

	err := http.ListenAndServe(listenOnAddrPort, &handler)
	if err != nil {
		log.Fatal(err)
	}
}

func loaderCmd(etcdEndpoints []string, generation string, k int, port int64) error {
	srvs, err := getMicroservicesList(etcdEndpoints)
	if err != nil {
		return err
	}

	if k == -1 {
		k = len(srvs)
	}
	chains := combin.Combinations(len(srvs), k)

	fmt.Print("chain,chain_length,duration\n")
	for _, indexes := range chains {
		chain := []string{}
		for _, idx := range indexes {
			chain = append(chain, srvs[idx])
		}

		//we get the payload through the chain and we measure how much time it takes
		start := time.Now()

		_, err := getPayloadFromChain(chain, port)

		duration := time.Since(start)

		if err != nil {
			log.Print(err)
		}

		fmt.Printf("\"%v\",%d,%d\n", chain, len(chain), duration)
	}

	return nil
}

func clearCmd(etcdEndpoints []string) error {
	return clearAllLeases(etcdEndpoints)
}

func main() {

	interfaceName := flag.String("i", "eth0", "interface name to read IP from. A single IP must be configured (defaults to 'eth0')")
	ipv6 := flag.Bool("v6", false, "(flag) Use ipv6 (defaults to false)")
	etcdEndpoints := flag.String("etcd", "http://127.0.0.1:2379", "Etcd endpoints comma separated (defaults to 'http://127.0.0.1:2379')")
	generation := flag.String("g", "0", "Used to identify between generations of microservices (defaults to '0')")
	port := flag.Int64("p", 3365, "Listen port (defaults to '3365')")
	ip := flag.String("a", "0.0.0.0", "Ip to bind on (defaults to '0.0.0.0')")
	payloadSize := flag.Int64("payloadSize", 64, "Payload size in bytes (defaults to 64)")
	k := flag.Int("k", -1, "Max chain len. (defaults to -1 - no limit)")

	flag.Parse()

	cmd := flag.Arg(0)

	if cmd == "" {
		fmt.Printf("Usage: %s [options] <cmd> \n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(-1)
	}

	etcdEndpointsArr := strings.Split(*etcdEndpoints, ",")

	switch cmd {
	case "server":

		err := serverCmd(*interfaceName, *ipv6, etcdEndpointsArr, *generation, *ip, *port, *payloadSize)
		if err != nil {
			log.Fatal(err)
		}
	case "loader":

		err := loaderCmd(etcdEndpointsArr, *generation, *k, *port)
		if err != nil {
			log.Fatal(err)
		}
	case "clear":
		err := clearCmd(etcdEndpointsArr)
		if err != nil {
			log.Fatal(err)
		}
	}
}
