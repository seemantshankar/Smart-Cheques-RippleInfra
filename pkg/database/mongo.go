package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoConnection(connectionString, databaseName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(databaseName)
	log.Printf("Successfully connected to MongoDB database: %s", databaseName)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Client.Ping(ctx, nil)
}

func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

func (m *MongoDB) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := m.Database.RunCommand(ctx, map[string]interface{}{"dbStats": 1}).Decode(&result)
	return result, err
}