package db

import (
	"context"

	"github.com/cz-theng/czkit-go/log"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	DBName               = "d_gametaverse"
	TransactionTableName = "t_transactions"
)

// CreateAndInitDB connect mongo and create db and collection

func CreateAndInitDB(URI string) (err error) {
	log.Info("DBURI:%v", URI)
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new mongo client error:%s", err.Error())
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Error("mongo connect error:%s", err.Error())
		return err
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("mongo ping error:%s", err.Error())
		return err
	} else {
		log.Info("connect db success")
	}

	db := client.Database("d_gametaverse")
	if db == nil {
		log.Error("db gametaverse is null")
		return err
	}

	gamesTbl := db.Collection("t_games")
	if gamesTbl != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := gamesTbl.Drop(ctx)
		if err != nil {
			log.Error("drop collection games transaction error:", err)
			return err
		}
	}
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = db.CreateCollection(ctx, "t_games")
	if err != nil {
		log.Error("create collection games error:", err)
		return err
	}

	trxTbl := db.Collection("t_transaction")
	if trxTbl != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := trxTbl.Drop(ctx)
		if err != nil {
			log.Error("drop collection games transaction error:", err)
			return err
		}
	}
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = db.CreateCollection(ctx, "t_transaction")
	if err != nil {
		log.Error("create collection games error:", err)
		return err
	}

	log.Info("done init mongon db!")
	return
}
