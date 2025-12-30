# Darda-chat

Scalable chat application (Go, MongoDB, Redis, RabbitMQ)

![Design of Darda-chat](https://github.com/user-attachments/assets/b8462dfa-34ce-4cdd-a648-738c107fa115)

## Design overview

Darda-chat's design follows a **decoupled, event-driven architecture** optimized for scalability and fault tolerance:

- **Gateway (Istio)**
  - Handles authentication, rate limiting, and load balancing
  - Routes traffic to services

- **WebSocket Servers**
  - Manage active client connections
  - Publish messages and user status updates
  - Support horizontal scaling using redis sub/pub
  - Handle users's multiple client connections using the same redis pub/sub channel (one channel at most per user) in linear time
  - Handle slow clients gracefully using buffered channels and terminate connection if client reaches buffer limit

- **Redis**
  - Pub/Sub for real-time message fan-out to enable WS servers horizontal scaling
  - Caching for messages, store online states

- **RabbitMQ**
  - Decouples real-time messaging from persistence

- **Messages Writer Service**
  - Consumes messages from RabbitMQ and persists messages to MongoDB
  - Notify read services about new messages by publishing Persistent Message events to RabbitMQ in a reliable way using outbox pattern

- **Messages Reader Service**
  - Loads chat history messages from MongoDB, with pagination supported
  - Updates message cache

- **MongoDB**
  - Stores users and chat messages
  - Runs as a replica set to support Transactions

This design allows:

- Independent scaling of WebSocket servers, reads and writes services
- Resilient message handling
- Eventual consistency with fast real-time delivery

## Prerequisites

- Kubernetes cluster
- Skaffold installed in your environment (for development only)

## Configs

Configuration is loaded from environment variables at startup of each Service. See the `README.md` of each service for concrete configs and defaults values, as well you can find the used configs in the Configmaps and Secrets in the `manifests` folder.

## Run (development)

This projects uses Skaffold, you can run it by running:

```bash
skaffold dev
# or if you're using a dev registry for images.
skaffold dev --default-repo=localhost:5000
# Port-forward the chat-service to access the
kubectl port-forward services/chat-service 8080:8080
# Now using the Html page you can connect and send/receive messages.
# For now there is no front-end yet, a simple html page is provided to interact with the backend, the auth is not yet implemented, we only rely on the ID parameter provided in the URL by the client.
```

### Extra step

For easier dev env, auth was disabled for MongoDB to be able to use replicaset mode without pain of keyfile generation. This is not recommended for production.

So we need to manually initialize the replicaset:

```js
rs.initiate({
  _id: "rs0",
  members: [
    { _id: 0, host: "mongo-0.mongo.mongodb.svc.cluster.local:27017" }
  ]
})
```

No production images are available for now, Skaffold will build and push the images.

## Roadmap

- [x] Add multi-server support for WS servers
- [x] Add Kubernetes deployment
- [x] Add Messages DB (MongoDB) and messages persistence (Async write using Message broker)
- [x] Add History loading when connected
- [x] Add messages caching, logic to handle lost messages when a user connect after message is sent to message queue.
- [ ] Switch to local message routing instead of one redis channel per user for better scalability
- [ ] Add seen, delivered, sent status
- [ ] Develop the front end
- [ ] Add Istio Service Mesh (Gateway)
- [ ] Add Authentication (Istio Gateway + OAuth2-Proxy + Gmail authentication)
- [ ] Add Observability and Monitoring (Prometheus, Grafana, Opentelemetry)
- [ ] Add notifications
- [ ] Improve reliability (retries, circuit breaker.. etc)
- [ ] Add end-to-end encryption
- [ ] Add Helm Charts
- [ ] Add support to enable/disable SSL verification
