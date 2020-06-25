package main

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var _ETCDKey = "/simulator_ips"

func registerInEtcd(key string, generation string, etcdEndpoints []string) {
	cfg := clientv3.Config{
		Endpoints: etcdEndpoints,
		// set timeout per request to fail fast when the target endpoint is unavailable
		DialTimeout: time.Second,
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = cli.Put(ctx, key, generation)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	// use the response

}
