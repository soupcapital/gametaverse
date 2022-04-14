package daily

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/gametaverse/gamefidata/api"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/utils"
	"github.com/gametaverse/gfdp/rpc/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	mngopts "go.mongodb.org/mongo-driver/mongo/options"
)

func NewDailyTask() *DailyTask {
	task := &DailyTask{}
	return task
}

type DailyTask struct {
	ctx       context.Context
	cancelFun context.CancelFunc
	dbClient  *mongo.Client
	db        *mongo.Database
	hauTbl    *mongo.Collection
	gameTbl   *mongo.Collection
	rpcUrl    string
}

func (task *DailyTask) Init(mongoUrl, rpcUrl string) (err error) {
	task.ctx, task.cancelFun = context.WithCancel(context.Background())
	task.rpcUrl = rpcUrl
	err = task.initDB(mongoUrl)
	if err != nil {
		log.Printf("Init mongon error:%s", err.Error())
		return err
	}
	return
}

func (task *DailyTask) QueryHau(day time.Time) (err error) {
	for _, c := range api.AllChain {
		hau, err := task.HauForChain(c, day)
		if err != nil {
			return err
		}
		log.Printf("hau for %v is %d", c, hau)
		cn := api.UnparsePBChain(c)
		hauData := &db.Hau{
			ID:        fmt.Sprintf("%s@%d", cn, day.Unix()),
			Timestamp: uint64(day.Unix()),
			Chain:     cn,
			Hau:       hau,
		}

		ctx, cancel := context.WithTimeout(task.ctx, 100*time.Second)
		defer cancel()

		opt := mngopts.Update()
		opt.SetUpsert(true)
		update := bson.M{
			"$set": bson.M{
				"ts":    hauData.Timestamp,
				"chain": hauData.Chain,
				"hau":   hauData.Hau,
			},
		}
		_, err = task.hauTbl.UpdateByID(ctx, hauData.ID, update, opt)
		if err != nil {
			log.Printf("Update  monitor latest error: %s", err.Error())
			continue
		}

	}
	return nil
}

func (task *DailyTask) HauForChain(chain pb.Chain, day time.Time) (hau int64, err error) {

	contracts, err := task.getAllContractsOfChain(chain)
	if err != nil {
		log.Printf("getAllContractsOfChain : %v", err)
		return
	}
	conn, err := grpc.Dial(task.rpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("RPC did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(task.ctx, 30000*time.Second)
	defer cancel()
	dauRsp, err := c.Dau(ctx, &pb.GameReq{
		Start:     utils.BeginTimestamp,
		End:       day.Unix(),
		Contracts: contracts,
	})
	if err != nil {
		log.Printf("Dau error:%s", err.Error())
		return
	}
	hau = int64(dauRsp.Dau)
	return
}

func (task *DailyTask) initDB(URI string) (err error) {
	task.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Printf("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(task.ctx, 10*time.Second)
	err = task.dbClient.Connect(ctx)
	if err != nil {
		log.Printf("connect mongo error:%s", err.Error())
		return
	}

	err = task.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("ping mongo error:%s", err.Error())
	} else {
		log.Printf("connect mongo success")
	}

	task.db = task.dbClient.Database(db.DBName)
	if task.db == nil {
		log.Printf("db is null, please init db first")
		return
	}

	task.gameTbl = task.db.Collection(db.GameInfoTableName)
	if task.gameTbl == nil {
		log.Printf("collection is null, please init db first")
		return
	}

	task.hauTbl = task.db.Collection(db.HauTableName)
	if task.hauTbl == nil {
		log.Printf("collection is null, please init db first")
		return
	}

	return
}

func (task *DailyTask) getAllContractsOfChain(chain pb.Chain) (contracts []*pb.Contract, err error) {
	if chain == pb.Chain_UNKNOWN {
		return
	}
	gameTbl := task.db.Collection(db.GameInfoTableName)

	ctx, cancel := context.WithTimeout(task.ctx, 10*time.Second)
	defer cancel()

	chainName := api.UnparsePBChain(chain)

	cursor, err := gameTbl.Find(ctx, bson.M{
		"chain": chainName,
	})
	if err != nil {
		log.Printf("Find game error: %v", err.Error())
		return
	}

	var games []db.Game
	if err = cursor.All(ctx, &games); err != nil {
		log.Printf("Find game error: %v", err.Error())
		return
	}

	for _, g := range games {
		cn := api.ParsePBChain(g.Chain)
		for _, c := range g.Contracts {
			pc := &pb.Contract{
				Chain:   cn,
				Address: strings.ToLower(c),
			}
			contracts = append(contracts, pc)
		}
	}
	return
}
