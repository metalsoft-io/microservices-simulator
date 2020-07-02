package main

import (
	"fmt"
	"log"
	"math/rand"
	mrand "math/rand"
	"time"
)

func loaderCmd(etcdEndpoints []string, generation string, maxChainLength int, port int64, count int, showChain bool, payloadSize int64, disableKeepAlive bool, timeout int) error {
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

		//we get the payload through the chain and we measure how much time it takes
		start := time.Now()

		chain := generateChain(srvs, k)

		_, err := getPayloadFromChain(chain, port, payloadSize, disableKeepAlive, timeout)

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

func generateRandomUniqueIntegers(length int) []int {
	arr := []int{}
	for {
		newI := rand.Intn(length)
		found := false
		for _, i := range arr {
			if newI == i {
				found = true
				break
			}
		}

		if !found {
			arr = append(arr, newI)
		}

		if len(arr) == length {
			break
		}
	}
	return arr
}

func generateChain(srvs []string, chainLength int) []string {

	if len(srvs) < chainLength {
		log.Fatal("srvs < chainLength")
	}

	indexes := generateRandomUniqueIntegers(chainLength)

	chain := []string{}
	for _, idx := range indexes {
		chain = append(chain, srvs[idx])
	}

	return chain
}
