# microservices-simulator


This application simulates a number of microservices comunicating with each other and measures accumulated latency.
It works by simulating a chain of microservices calls (one microserice calls another, that calls another ...) and then the end of the chain returns the requersted (random) payload to the original requester via the chain. This showcases latency amplification which depends on the cluster topology and networking used.

The simulator can vary the chain depth, disablekeepalive etc.

## Starting the cluster
BY default this starts 40 microservices clients:
```bash
kubectl apply -f microservices-simulator.yml
```


## Executing the benchmark
```bash
kubectl run -it --rm --restart=Never ms --image=metalsoft/microservices-simulator --command microservices-simulator -- -etcd http://etcd-client:2379  -k 0 -n 1000 -payloadSize 64 loader > results.csv
```

## Arguments
```
microservices-simulator <cmd> [arguments]

server: starts a server node (part of a cluster)
	-i interface name to read IP from. A single IP must be configured (defaults to 'eth0')
	-v6 (flag) Use ipv6 (defaults to false)
	-etcd Etcd endpoints comma separated (defaults to 'http://127.0.0.1:2379')
	-g Used to identify between generations of microservices (defaults to '0')
	-p Listen port (defaults to '3365')
	-a Ip to bind on (defaults to '0.0.0.0') 

loader: executes a benchmark run against a cluster
	--payloadSize Payload size in bytes (defaults to 64)
	-k Max chain len. (defaults to -1 - no limit). Set to zero to choose a random number every time.	
	-n test count. (defaults to 1)	
	--showChain  Show chain (defaults to false)
	--disableKeepalives If set disables connection reuse. Defaults to false (enabled).	
	--timeout HTTP client timeout in seconds. (defaults to 10s)	

clear:
	Erases the current configuration from etcd
```


## Interpreting results
The loader command will return a csv containing:
`chain_len`, `duration_in_seconds`

```
10,222
11,111
....
```

If `--showChain` is used the format will be `chain,chain_len,duration_in_seconds`
```
"192.168.10.10,192.168.10.11",2,100 
....
```


## Requirements:

An etcd cluster is required for the microservices to discover each other. 
To deploy an etcd cluster in kubernetes use the following manifest:

```bash
kubectl apply -f https://raw.githubusercontent.com/etcd-io/etcd/master/hack/kubernetes-deploy/etcd.yml
```

