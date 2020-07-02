package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

	disableKeepalives := flag.Bool("disableKeepalives", false, "If set disables connection reuse. Defaults to false (enabled).")
	timeout := flag.Int("timeout", 10, "HTTP client timeout in seconds. (defaults to 10s)")

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

		err := serverCmd(*interfaceName, *ipv6, etcdEndpointsArr, *generation, *ip, *port, *disableKeepalives, *timeout)
		if err != nil {
			log.Fatal(err)
		}
	case "loader":

		err := loaderCmd(etcdEndpointsArr, *generation, *k, *port, *n, *showChain, *payloadSize, *disableKeepalives, *timeout)
		if err != nil {
			log.Fatal(err)
		}
	case "clear":
		err := clearAllLeases(etcdEndpointsArr)
		if err != nil {
			log.Fatal(err)
		}
	}
}
