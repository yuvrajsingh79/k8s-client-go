# k8s-client-go
Custom Go client for interacting with k8s cluster using gRPC (http2 protocol). 

## Pre-Requisites

* Go version 1.14.xx or above must be installed
* k8s version 1.18.xx shou;d be installed 
* Docker should be installed
* If running over Minikube then its version should be 1.11.0 or above


## Usage

```bash
git clone https://github.com/yuvrajsingh79/k8s-client-go.git
cd k8s-client-go/
go run main.go

cd grpc-server/
go run main.go
```

Now, we are to test our code, head over to postman and perform the follwing GET api requests :
* Get Pods and Services running in the cluster ```localhost:2000/getPodServ```
* Create a custom resource and its defination, it also deploys a service using load balancer   ```localhost:2000/createcrd```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
