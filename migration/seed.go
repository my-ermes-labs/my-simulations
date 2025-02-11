package main

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/ermes-labs/api-go/api"
	"github.com/ermes-labs/api-go/infrastructure"
	rc "github.com/ermes-labs/storage-redis/packages/go"
	"github.com/redis/go-redis/v9"
)

// The node that the function is running on.
var node *api.Node

// The Redis client.
var redisClient *redis.Client

func seed() {
	// Initialize a new Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // replace with your Redis server address
		Password: "",               // replace with your password if needed
		DB:       0,                // default DB
	})

	// Define the context for the Redis operations
	ctx := context.Background()

	// Define the keys and their sizes in KB
	keys := []int{1, 256, 512, 1024, 2048, 3072, 4096}

	// Iterate over the keys
	for _, key := range keys {
		// Create a string that weights as much as the key in KB
		value := strings.Repeat("a", key*1024) // each character is 1 byte, so multiply by 1024 to get KB

		// Create the keyspace.
		ks := rc.NewErmesKeySpaces(strconv.Itoa(key))
		// Insert the key-value pair into the Redis database
		err := rdb.Set(ctx, ks.Session("key"), value, 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func init() {
	// Get the node from the environment variable.
	jsonNode := os.Getenv("ERMES_NODE")
	// Unmarshal the environment variable to get the node.
	infraNode, err := infrastructure.UnmarshalNode([]byte(jsonNode))
	// Check if there was an error unmarshalling the node.
	if err != nil {
		panic(err)
	}

	// Get the Redis connection details from the environment variables.
	redisHost := envOrDefault("REDIS_HOST", "localhost")
	redisPort := envOrDefault("REDIS_PORT", "6379")
	redisPassword := envOrDefault("REDIS_PASSWORD", "")
	// Create a new Redis client.
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0, // use default DB
	})

	// The Redis commands.
	var RedisCommands = rc.NewRedisCommands(redisClient)
	// Create a new node with the Redis commands.
	node = api.NewNode(*infraNode, RedisCommands)
}

// Get the value of an environment variable or return a default value.
func envOrDefault(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}

// Get the value of an environment variable or panic if it is not set.
func envOrPanic(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(key + " env variable is not set")
	}
	return value
}
