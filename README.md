# Darda-chat

Scalable chat application (Go, MongoDB, Redis, RabbitMQ)

![Design of Darda-chat](https://github.com/user-attachments/assets/c0976fa7-8e4b-4b57-8ad0-b2a402a8330d)

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

No production images are available for now, Skaffold will build and push the images.

## Roadmap

- [x] Add multi-server support for WS servers
- [x] Add Kubernetes deployment
- [ ] Add Messages DB (MongoDB) and messages persistence (Async write using Message broker)
- [ ] Add History loading when connected
- [ ] Add messages caching
- [ ] Add seen, delivered, sent status
- [ ] Improve reliability (retries, circuit breaker.. etc)
- [ ] Add Istio Service Mesh (Gateway)
- [ ] Add support to enable/disable SSL verification
- [ ] Develop the front end
- [ ] Add Helm Charts
