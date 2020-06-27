package main

import (
	"flag"
	"fmt"
	"log"
	mrand "math/rand"
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

func serverCmd(interfaceName string, ipv6 bool, etcdEndpoints []string, generation string, listenOnIP string, port int64) error {
	//determine my ip
	myIP := getMyIP(interfaceName, ipv6)
	log.Printf("Discovered IP %s from interface %s", myIP, interfaceName)

	finished := make(chan bool)
	log.Printf("Registering %s in etcd %v", myIP, etcdEndpoints)

	//register myself in etcd and start the renew lease loop
	go registerInEtcdAndRenewLeases(myIP, port, etcdEndpoints, finished)

	//start http server
	go startHTTPServer(listenOnIP, port, myIP)

	//forever loop of the main thread
	for {
		time.Sleep(5 * time.Second)
	}

}

func startHTTPServer(listenOnIP string, port int64, myIP string) {

	handler := myHandler{
		myIP: myIP,
	}

	handler.Port = port

	listenOnAddrPort := fmt.Sprintf("%s:%d", listenOnIP, port)
	log.Printf("Listening on %s", listenOnAddrPort)

	err := http.ListenAndServe(listenOnAddrPort, &handler)
	if err != nil {
		log.Fatal(err)
	}
}

func loaderCmd(etcdEndpoints []string, generation string, maxChainLength int, port int64, count int, showChain bool, payloadSize int64) error {
	srvs, err := getMicroservicesList(etcdEndpoints)
	if err != nil {
		return err
	}

	if showChain {
		fmt.Print("chain,chain_length,duration\n")
	} else {
		fmt.Print("chain_length,duration\n")
	}

	mrand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {

		k := maxChainLength
		switch maxChainLength {
		case -1:
			k = len(srvs)
		case 0:
			k = mrand.Intn(len(srvs)) + 1
		}

		chains := combin.Combinations(len(srvs), k)

		indexes := chains[0]

		//	for _, indexes := range chains {

		chain := []string{}
		for _, idx := range shuffle(indexes) {
			chain = append(chain, srvs[idx])
		}

		//we get the payload through the chain and we measure how much time it takes
		start := time.Now()

		_, err := getPayloadFromChain(chain, port, payloadSize)

		duration := time.Since(start).Seconds()

		if err != nil {
			duration = -1
		}

		if showChain {
			fmt.Printf("\"%v\",%d,%f\n", chain, len(chain), duration)
		} else {
			fmt.Printf("%d,%f\n", len(chain), duration)
		}

		//	}
	}

	return nil
}

func shuffle(a []int) []int {
	mrand.Seed(time.Now().UnixNano())
	mrand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
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
	k := flag.Int("k", -1, "Max chain len. (defaults to -1 - no limit). Set to zero to choose a random number every time.")
	n := flag.Int("n", 1, "test count. (defaults to 1)")
	showChain := flag.Bool("showChain", false, "Show chain (defaults to false)")

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

		err := serverCmd(*interfaceName, *ipv6, etcdEndpointsArr, *generation, *ip, *port)
		if err != nil {
			log.Fatal(err)
		}
	case "loader":

		err := loaderCmd(etcdEndpointsArr, *generation, *k, *port, *n, *showChain, *payloadSize)
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
