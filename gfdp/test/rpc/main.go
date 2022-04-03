package main

import (
	"context"
	"log"
	"time"

	"github.com/gametaverse/gfdp/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = "172.31.6.11:8082"
)

func main() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gameOne := []*pb.Contract{}
	gameOne = append(gameOne, &pb.Contract{Chain: pb.Chain_BSC, Address: "0xe9e7cea3dedca5984780bafc599bd69add087d56"})
	gameOne = append(gameOne, &pb.Contract{Chain: pb.Chain_BSC, Address: "0xe9e7cea3dedca5984780bafc599bd69add087d57"})
	gameOne = append(gameOne, &pb.Contract{Chain: pb.Chain_POLYGON, Address: "0xe9e7cea3dedca5984780bafc599bd69add087d57"})

	gameTwo := []*pb.Contract{}
	gameTwo = append(gameTwo, &pb.Contract{Chain: pb.Chain_BSC, Address: "0x8c851d1a123ff703bd1f9dabe631b69902df5f97"})
	gameTwo = append(gameTwo, &pb.Contract{Chain: pb.Chain_BSC, Address: "0x8c851d1a123ff703bd1f9dabe631b69902df5f98"})
	gameTwo = append(gameTwo, &pb.Contract{Chain: pb.Chain_POLYGON, Address: "0x8c851d1a123ff703bd1f9dabe631b69902df5f98"})

	r, err := c.TwoGamesPlayers(ctx, &pb.TwoGamesPlayersReq{
		Start:   1648624801,
		End:     1648711201,
		GameOne: gameOne,
		GameTwo: gameTwo,
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Dau: %v", r.Users)

}

func main_2() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var users []*pb.Contract
	users = append(users, &pb.Contract{
		Chain:   pb.Chain_BSC,
		Address: "0x8894e0a0c962cb723c1976a4421c95949be2d4e3",
	})
	users = append(users, &pb.Contract{
		Chain:   pb.Chain_POLYGON,
		Address: "0x8894e0a0c962cb723c1976a4421c95949be2d4e3",
	})
	r, err := c.AllUserPrograms(ctx, &pb.AllUserProgramsReq{
		Start: 1646784000,
		End:   1646870400,
		Users: users,
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Dau: %v", r.Programs)

}

func main_1() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDBProxyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := c.Dau(ctx, &pb.GameReq{
		Start: 1646784000,
		End:   1646870400,
		Contracts: []*pb.Contract{
			{
				Address: "0x5b7d8a53e63f1817b68d40dc997cb7394db0ff1a",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0xf35aee66ab0d099af233c1ab23e5f2138b0ed15b",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0x370ce09af3ee5e0e3f9f7f3b661505d1fbdc6ec6",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0x1021a5ac2fff0f9fe4cf8f877ad2748f61defa06",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0x6f9982f5213c6e3d8b130fe031d396963a1af5b5",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0xa3b4d483683e838ca7013d576e09cac59b839325",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0xd63bce6a1eea0cdd5f79489551010a2e355e5f71",
				Chain:   pb.Chain_BSC,
			},
		},
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Dau: %v", r.Dau)

	r2, err := c.TxCount(ctx, &pb.GameReq{
		Start: 1646784000,
		End:   1646870400,
		Contracts: []*pb.Contract{
			{
				Address: "0x5b7d8a53e63f1817b68d40dc997cb7394db0ff1a",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0xf35aee66ab0d099af233c1ab23e5f2138b0ed15b",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0x370ce09af3ee5e0e3f9f7f3b661505d1fbdc6ec6",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0x1021a5ac2fff0f9fe4cf8f877ad2748f61defa06",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0x6f9982f5213c6e3d8b130fe031d396963a1af5b5",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0xa3b4d483683e838ca7013d576e09cac59b839325",
				Chain:   pb.Chain_BSC,
			},
			{
				Address: "0xd63bce6a1eea0cdd5f79489551010a2e355e5f71",
				Chain:   pb.Chain_BSC,
			},
		},
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("TxCount: %v", r2.Count)
}
