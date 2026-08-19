package Init

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"log"
	"os"
)

type MilvusManger struct {
	Client *milvusclient.Client
}

func NewMilvusManger() *MilvusManger {
	if err := godotenv.Load(); err != nil {
		log.Printf("skip loading .env: %v", err)
	}
	ctx := context.Background()
	newClient, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: os.Getenv("address"),
		APIKey:  os.Getenv("apikey"),
	})
	if err != nil {
		panic(err)
	}
	return &MilvusManger{
		Client: newClient,
	}
}
