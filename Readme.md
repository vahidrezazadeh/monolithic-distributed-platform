
# Monolithic Distributed Platform with Go, Gin, and Docker
This project implements a multi-instance monolithic distributed platform using Go, Gin, and Docker. The architecture combines the simplicity of a monolithic design with distributed system capabilities, featuring:

## Features ✨
- Horizontal scaling with multiple instances
- Service discovery and cluster management
- Distributed job processing
- Redis for caching and distributed locking
- Kafka for event streaming
- Health checks and monitoring endpoints
- Leader election

## Prerequisites
- Docker 20.10+
- Docker Compose 1.29+
- Golang 1.24+
- Internet (In Iran)

# Quick Start

Clone the repository:
```
git clone https://github.com/vahidrezazadeh/monolithic-distributed-platform.git
cd monolithic-distributed-platform
```
Start the system:
```
docker-compose up -d
```
## Architecture
The system consists of :
- 3 application instances (can be scaled)
- Redis for shared state
- Kafka for event streaming
- ZooKeeper for Kafka coordination

## API Endpoints
All instances expose the same API:
|Endpoint | Method | Description
| ------- | ------ | -----------
|/health | GET | Health check|
|/nodes|	GET	|List cluster nodes|
|/jobs|	POST|	Submit a new job|

## Configuration
Environment variables for application instances:
| Variable        | Default          | Description                          |
|------------------|------------------|--------------------------------------|
| INSTANCE_ID      | auto-generated    | Unique instance identifier           |
| HTTP_PORT        | 8080             | HTTP API port                        |
| GRPC_PORT        | 50051            | gRPC port                            |
| CLUSTER_PORT     | 7946             | Cluster communication port           |
| JOIN_ADDRESSES   | ""               | Comma-separated cluster addresses    |
| REDIS_ADDRESS     | "redis:6379"     | Redis server address                 |
| KAFKA_ADDRESS    | "kafka:9092"     | Kafka broker address                 |
| WORKER_COUNT     | 10               | Number of job workers per instance   |
| MAX_QUEUE_SIZE| 1000 | Maximum job queue size|

## Scaling
To scale the application instances:

```bash
docker-compose up -d --scale app1=1 --scale app2=3 --scale app3=2
```
This will create:

- 1 instance of app1 (seed node)
- 3 instances of app2
- 2 instances of app3

## Monitoring

Check individual instance health:
```bash
curl http://localhost:8080/api/v1/health
```
View cluster nodes:

```bash
curl http://localhost:8080/api/v1/nodes
```

## Clean Up
To stop and remove all containers:
```bash
docker-compose down
```

```bash
docker-compose down -v
```

## 🔗 Links
[![portfolio](https://img.shields.io/badge/my_portfolio-000?style=for-the-badge&logo=ko-fi&logoColor=white)](https://vahidrezazadeh.github.io/)
[![linkedin](https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/vahidrezazadeh/)
[![twitter](https://img.shields.io/badge/twitter-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white)](https://twitter.com/vahdirezazadeh5)