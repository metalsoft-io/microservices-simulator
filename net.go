package main

import (
	"log"
	"net"
)

//getMyIP returns the first ip configured on the specified interface. if ipv6 is set to true
func getMyIP(interfaceName string, ipv6 bool) string {

	log.Printf("Interrogating interface %s for ip...", interfaceName)

	intf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := intf.Addrs()

	if err != nil {
		log.Fatal(err)
	}

	if len(addrs) == 0 {
		log.Fatalf("interface %s has no ips", interfaceName)
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			log.Printf("Error parsing ip:%v", err)
			continue
		}

		if !ipv6 && ip.To4() != nil {
			return ip.String()
		} else if ipv6 && ip.To4() == nil {
			return ip.String()
		}
	}

	log.Fatalf("Could not determine ip for interface %s", interfaceName)

	return ""
}
