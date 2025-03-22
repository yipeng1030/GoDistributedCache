# GoDistributedCache

A distributed, cloud-native caching system implemented in Go.  
Supports dynamic scaling, hot-key replication, and cache-safe mechanisms powered by Kubernetes.

## ✨ Features

- **Distributed Communication**  
  Implements a peer-to-peer caching protocol using HTTP + Protobuf for efficient inter-node communication.

- **Thread-safe Caching Core**  
  Uses concurrency-safe **LRU** as default with pluggable support for **FIFO** and **LFU** eviction strategies.

- **Two-tier Caching with Hot-key Replication**  
  Introduces hot key mirroring between nodes to reduce cross-node network overhead.

- **Cache Safety Mechanisms**
    - **Consistent Hashing**: Ensures stable key routing, minimizes cache invalidation when scaling.
    - **SingleFlight**: Prevents cache breakdown by deduplicating concurrent requests for the same key.

- **Cloud-native Deployment**
    - Uses Kubernetes Headless Service for automatic peer discovery.
    - Exposes services via \`Service\` with \`NodePort\` support.
    - Enables one-command rolling update & auto-scaling via \`Deployment\`.

## 🛠️ Tech Stack

- **Language**: Go 1.22+
- **Protocol**: HTTP + Protobuf
- **Container**: Docker
- **Orchestration**: Kubernetes (Minikube tested)
- **Architecture**: Peer-to-peer; each node is both server & client

## 📁 Project Structure

```
├── main/                   # Entry point, API + DNS-based peer discovery  
├── consistenthash/         # Consistent hashing logic  
├── singleflight/           # In-flight request deduplication  
├── obsolescence/           # LRU, LFU, FIFO eviction algorithms  
├── cachepb/                # Protobuf definition & generated Go code  
├── deploy/                 # Kubernetes YAML configs  
├── http.go                 # HTTP peer pool implementation  
├── peers.go                # PeerPicker interfaces  
├── cache.go                # Cache group & get logic  
├── byte_view.go            # ByteView abstraction  
```

## 🚀 Quick Start

### 1. Build the Docker Image

```bash
docker build --platform=linux/arm64 -t mycache:latest .
```

### 2. Load into Minikube

```bash
minikube image load mycache:latest
```

### 3. Deploy to Kubernetes

```bash
kubectl apply -f deploy/headless_service.yaml  
kubectl apply -f deploy/service.yaml  
kubectl apply -f deploy/deployment.yaml
```

### 4. Access the Cache API

```bash
minikube service mycache-service
```

## 📌 Design Highlights

- All nodes act as **peer-aware servers**—no external registry needed.
- Automatic peer updates via **DNS polling** from Headless Service.
- Dynamic pod scaling instantly updates peer list—ideal for elastic environments.
- Seamless **multi-node load balancing** with Kubernetes \`Service\`.

## 🤯 Challenges Overcome

> Designing and deploying a state-aware distributed cache system on Kubernetes is non-trivial.

- Solved **peer discovery** in dynamic cluster conditions using DNS-based polling.
- Built **fail-tolerant** data fetching with \`SingleFlight\` and **hot-key mirroring**.
- Ensured safe cache routing & rebalancing with **consistent hashing**.
- Achieved one-command deployment & scaling via \`Deployment\` & Docker image packaging.

## 🧪 Testing

Unit tests for eviction logic, consistent hash ring, and peer selection can be found under:

```
consistenthash/consistenthash_test.go  
obsolescence/obsolescene_test.go  
go_distribute_cache_test.go  
```

## 📎 TODO

- [ ] gRPC support for higher performance
- [ ] Prometheus metrics endpoint
- [ ] Support cache expiration & TTL
- [ ] CI/CD integration with GitHub Actions

## 📄 License

MIT License © 2025 mengyipeng