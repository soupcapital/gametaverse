package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// CreateAndInitDB connect mongo and create db and collection
func CreateAndInitDB(username, passwd string) (err error) {
	URI := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.g9w77.mongodb.net/?retryWrites=true&w=majority", username, passwd)
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("connect db success")
	}

	db := client.Database("db_cti")
	if db == nil {
		log.Fatal("db cti is null")
	}
	log.Println("got db db_cti")

	tweetsTbl := db.Collection("tweets")
	if tweetsTbl == nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = db.CreateCollection(ctx, "tweets")
		if err != nil {
			log.Fatal("create collection tweets error:", err)
			return
		}
		log.Println("done create collection")
	}

	log.Println("done init mongon db!")
	return
}
