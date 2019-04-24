package main

import (
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"os"

	pb "github.com/farerpath/albumservice/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"go.mongodb.org/mongo-driver/mongo"
)

type albumService struct{}

const port = ":17080"

var (
	DBADDR = "mongodb://maindb-service:27017"
)

const (
	ALBUMLISTDB = "albumlists"
	ALBUMDB     = "albums"
	PICTUREDB   = "pictures"
)

var (
	albumCollection     *mongo.Collection
	albumListCollection *mongo.Collection
	pictureCollection   *mongo.Collection
)

func init() {
	if addr := os.Getenv("FP_DB_ADDRESS"); len(addr) > 1 {
		log.Printf("DB address received: %v\n", addr)
		DBADDR = addr
	}
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(DBADDR))
	if err != nil {
		log.Fatalf("DB Connection failed:\n%v", err)
	}

	err = client.Connect(nil)
	if err != nil {
		log.Fatalf("DB Connection failed:\n%v", err)
	}

	albumCollection = client.Database("farerpath").Collection(ALBUMDB)
	albumListCollection = client.Database("farerpath").Collection(ALBUMLISTDB)
	pictureCollection = client.Database("farerpath").Collection(PICTUREDB)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Failed to listen tcp %v", err)
	}

	g := grpc.NewServer()

	pb.RegisterAlbumServer(g, &albumService{})

	reflection.Register(g)

	err = g.Serve(listener)
	if err != nil {
		log.Printf("Failed to start grpc server %v", err)
	}
}
