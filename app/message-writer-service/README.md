# Message Writer Service

The Message Writer Service is a microservice responsible for asynchronously writing messages to a database from a message broker (RabbitMQ). It provides a robust and scalable solution for handling real-time messaging scenarios.

## Configuration Variables

The following configuration variables can be set using environment variables:
* `AMQP_USER`: Username for RabbitMQ authentication. Default value: empty string.
* `AMQP_PASS`: Password for RabbitMQ authentication. Default value: empty string.
* `AMQP_HOST`: Host address of the RabbitMQ server. Default value: empty string.
* `MSG_QUEUE`: Name of the message queue to consume from. Default value: "messages".
* `MONGO_DB_NAME`: Name of the MongoDB database to write messages to. Default value: "darda_chat".
* `MONGO_COLLECTION_NAME`: Name of the MongoDB collection to write messages to. Default value: "messages".
* `MONGO_ADDR`: Address of the MongoDB server. Default value: "mongodb://localhost:27017".
* `MONGO_USER`: Username for MongoDB authentication. Default value: "root".
* `MONGO_PASS`: Password for MongoDB authentication. Default value: empty string.
* `MONGO_TIMEOUT`: Connection timeout for MongoDB. Default value: "10s".
* `REDIS_ADDR`: Address of the Redis server. Default value: "localhost:6379".
* `REDIS_PASS`: Password for Redis authentication. Default value: empty string.
* `REDIS_DB`: Database number to use in Redis. Default value: 0.
* `CONSUMER_POOL_SIZE`: Number of worker goroutines to consume messages. Default value: 10.

## Dependencies

The Message Writer Service depends on the following packages:
* github.com/nsmsb/darda-chat/app/message-writer-service/internal/config: Provides configuration variables and utility functions.
* github.com/nsmsb/darda-chat/app/message-writer-service/internal/consumer: Consumes messages from RabbitMQ and writes them to the database.
* github.com/nsmsb/darda-chat/app/message-writer-service/internal/db: Provides a MongoDB client for interacting with the database.
* github.com/nsmsb/darda-chat/app/message-writer-service/internal/handler: Defines message handlers for writing messages to the database.
* github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger: Provides a logger for logging application events.
* github.com/nsmsb/darda-chat/app/message-writer-service/pkg/rabbitmq: Provides utilities for interacting with RabbitMQ.
* github.com/redis/go-redis/v9: Provides a Redis client for interacting with the cache.
* go.uber.org/zap: Provides a structured logger for logging application events.

## Usage

Check the deployment manifests in `manifests/message-writer-service`
