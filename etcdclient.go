package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var _ETCDKey = "/simulator_ips"
var _LeaseDurationSeconds int64 = 5

func registerInEtcdAndRenewLeases(ip string, port int64, etcdEndpoints []string, finished chan bool) {
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

	kv := clientv3.NewKV(cli)

	key := fmt.Sprintf("%s/%s-%d", _ETCDKey, ip, port)
	url := fmt.Sprintf("http://%s:%d/", ip, port)

	renewLease(key, url, cli, kv)

	watchExpiredLease(key, url, cli, kv)

	finished <- true

}

func renewLease(key string, value string, cli *clientv3.Client, kv clientv3.KV) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	lease, _ := cli.Grant(ctx, _LeaseDurationSeconds)

	_, err := kv.Put(ctx, key, value, clientv3.WithLease(lease.ID))

	cancel()
	if err != nil {
		return err
	}

	return nil
}

func watchExpiredLease(key string, value string, cli *clientv3.Client, kv clientv3.KV) {

	prefix := fmt.Sprintf("%s", _ETCDKey)

	wCh := cli.Watch(context.Background(), prefix, clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithFilterPut())

	go func() {

		for wResp := range wCh {

			for _, ev := range wResp.Events {

				expired, err := isExpired(cli, ev)
				if err != nil {
					log.Println("Error when checking expiry")
				} else if expired {
					renewLease(key, value, cli, kv)
				}
			}
		}
	}()
}

// isExpired decides if a DELETE event happended because of a lease expiry
func isExpired(cli *clientv3.Client, ev *clientv3.Event) (bool, error) {
	if ev.PrevKv == nil {
		return false, nil
	}

	leaseID := clientv3.LeaseID(ev.PrevKv.Lease)
	if leaseID == clientv3.NoLease {
		return false, nil
	}

	ttlResponse, err := cli.TimeToLive(context.Background(), leaseID)
	if err != nil {
		return false, err
	}

	return ttlResponse.TTL == -1, nil
}

//retrieves a list of ips in the prefix
func getMicroservicesList(etcdEndpoints []string) ([]string, error) {

	cfg := clientv3.Config{
		Endpoints: etcdEndpoints,
		// set timeout per request to fail fast when the target endpoint is unavailable
		DialTimeout: time.Second,
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	defer cli.Close()

	prefix := fmt.Sprintf("%s/", _ETCDKey)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ret, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}

	srvs := []string{}
	for _, v := range ret.Kvs {
		srvs = append(srvs, string(v.Value))
	}

	return srvs, nil
}

func clearAllLeases(etcdEndpoints []string) error {
	cfg := clientv3.Config{
		Endpoints: etcdEndpoints,
		// set timeout per request to fail fast when the target endpoint is unavailable
		DialTimeout: time.Second,
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}

	defer cli.Close()

	prefix := fmt.Sprintf("%s/", _ETCDKey)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = cli.Delete(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return err
	}

	return nil

}
