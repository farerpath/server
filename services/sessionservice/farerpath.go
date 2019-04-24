package main

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/farerpath/sessionservice/proto"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/farerpath/server/model/model"

	jwt "github.com/dgrijalva/jwt-go"
)

// grpc listen port
const port = ":17080"

// DBADDRESS : MongoDB Address for store session
var DBADDRESS = "mongodb://localhost:27017"

// REDISADDR : Redis Address
var REDISADDR = "localhost:6379"

type sessionService struct{}

var (
	rClient *redis.Client
	mClient *mongo.Client
)

type FPClaim struct {
	Key 	string
	Type 	string
	jwt.StandardClaims
}

func init() {
	if addr := os.Getenv("FP_DB_ADDRESS"); len(addr)>1 {
		log.Printf("DB address received: %v\n", addr)
		DBADDRESS = addr
	}

	if addr := os.Getenv("FP_REDIS_ADDRESS"); len(addr)>1 {
		log.Printf("REDIS address received: %v\n", addr)
		REDISADDR = addr
	}
}

func main() {
	// Connect to Redis
	rClient = redis.NewClient(&redis.Options{
		Addr:     REDISADDR,
		Password: "",
		DB:       0,
	})
	defer rClient.Close()

	var err error
	// Connect to MongoDB
	mClient, err = mongo.NewClient(DBADDRESS)
	if err != nil {
		log.Fatalf("DB Connection failed:\n%v", err)
	}
	
	err = mClient.Connect(nil)
	if err != nil {
		log.Fatalf("DB Connection failed:\n%v", err)
	}

	log.Println("DB Connection succeed")

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen tcp:\n%v\n", err)
	}
	log.Printf("Listening at %v\n", port)

	g := grpc.NewServer()

	pb.RegisterSessionServer(g, &sessionService{})

	reflection.Register(g)

	err = g.Serve(listener)
	if err != nil {
		log.Printf("Failed to start grpc server %v", err)
	}
}

func (s *sessionService) NewSession(ctx context.Context, in *pb.SessionRequest) (*pb.SessionResponse, error) {
	v := secureRandomString(16)
	sign := secureRandomString(16)
	key := secureRandomString(32)

	authClaim := FPClaim{
		key,
		"AUTH",
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(in.SessionDurationTime)).Unix(),
			Issuer:    "farerpath",
		},
	}

	authToken := jwt.NewWithClaims(jwt.SigningMethodHS512, authClaim)
	authTokenString, err := authToken.SignedString([]byte(v))
	if err != nil {
		log.Printf("Failed to generate token.\n%v", err)
		return &pb.SessionResponse{}, status.Error(codes.Internal, err.Error())
	}

	refreshClaim := FPClaim {
		key,
		"REFRESH",
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(24*30)).Unix(),
			Issuer:    "farerpath",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaim)
	refreshTokenString, err := refreshToken.SignedString([]byte(sign))
	if err != nil {
		log.Printf("Failed to generate token.\n%v", err)
		return &pb.SessionResponse{}, status.Error(codes.Internal, err.Error())
	}

	// IP country query
	countryCode := ""

	ip := net.ParseIP(in.GetLoginIP())
	if ip != nil {
		// Check IP Location
		resp, err := http.Get("http://api.geoify.info/" + ip.String())
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		buffer, _ := ioutil.ReadAll(resp.Body)

		georesp := make(map[string]string)

		err = json.Unmarshal(buffer, georesp)
		if err != nil {
			log.Println(err)
		}

		c, exists := georesp["isoCountryCode"]
		if !exists {
			log.Println("err country code missing!")
		}

		countryCode = c
	}

	session := &model.UserSession{
		UserID:      	in.UserID,
		AuthToken:   	authTokenString,
		RefreshToken: 	refreshTokenString,
		LoginTime:   	time.Now().UTC(),
		DeviceType:  	in.LoginDeviceType,
		LoginIP:     	in.LoginIP,
		LoginRegion: 	countryCode,
		Sign:			sign,
	}

	collection := mClient.Database("farerpath").Collection("sessions")

	_, err = collection.InsertOne(ctx, session)
	// TODO: HAVE TO ROLLBACK ALL
	if err != nil {
		log.Printf("DB insertion failed,\n%v", err)
		return &pb.SessionResponse{}, status.Error(codes.Internal, err.Error())
	}

	rClient.Set(authTokenString, v, time.Hour*time.Duration(in.SessionDurationTime))

	return &pb.SessionResponse{AuthToken: authTokenString, RefreshToken:refreshTokenString}, nil
}

// TODO: extend
func (s *sessionService) ExtendSession(ctx context.Context, in *pb.ExtendRequest) (*pb.ExtendResponse, error) {
	collection := mClient.Database("farerpath").Collection("sessions")
	result := collection.FindOne(ctx, bson.D{{"userID", in.UserID}, {"refreshToken", in.RefreshToken}, {"authToken", in.AuthToken}})
	if result.Err() != nil {
		return &pb.ExtendResponse{}, status.Error(codes.Internal, result.Err().Error())
	}

	userSession := &model.UserSession{}
	result.Decode(userSession)

	t, err := jwt.ParseWithClaims(in.RefreshToken, &FPClaim{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(userSession.Sign), nil
	})
	if err != nil {
		log.Printf("jwt parsing error,\n%v", err)
		return &pb.ExtendResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	claim := t.Claims.(*FPClaim)
	
	if !t.Valid || claim.Type != "REFRESH" {
		return &pb.ExtendResponse{}, status.Error(codes.Unauthenticated, "")
	}
	
	authClaim := FPClaim{
		claim.Key,
		"AUTH",
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(in.SessionDurationTime)).Unix(),
			Issuer:    "farerpath",
		},
	}

	v, err := rClient.Get(in.AuthToken).Result()
	if err != nil {
		if err == redis.Nil {
			return &pb.ExtendResponse{}, status.Error(codes.InvalidArgument, "authToken already expired")
		}

		return &pb.ExtendResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	authToken := jwt.NewWithClaims(jwt.SigningMethodHS512, authClaim)
	authTokenString, err := authToken.SignedString([]byte(v))
	if err != nil {
		log.Printf("Failed to generate token.\n%v", err)
		return &pb.ExtendResponse{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.ExtendResponse{RefreshToken: in.RefreshToken, AuthToken: authTokenString}, nil
}

func (s *sessionService) GetAllSessions(ctx context.Context, in *pb.UserId) (*pb.UserSessions, error) {
	collection := mClient.Database("farerpath").Collection("sessions")
	cursor, err := collection.Find(ctx, bson.D{{"userID", in.UserID}})
	if err != nil {
		log.Printf("Find failed,\n%v", err)
		return &pb.UserSessions{}, status.Error(codes.Internal, err.Error())
	}

	var ns []*pb.UserSession

	for cursor.Next(ctx) {
		elem := &pb.UserSession{}
		cursor.Decode(elem)
		ns = append(ns, elem)
	}

	return &pb.UserSessions{Session: ns}, nil
}

func (s *sessionService) VerifyToken(ctx context.Context, in *pb.Token) (*pb.VerifyResponse, error) {
	v, err := rClient.Get(in.GetToken()).Result()
	if err != nil {
		if err == redis.Nil {
			return &pb.VerifyResponse{UserID: "", Valid: false}, nil
		}

		return &pb.VerifyResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	t, err := jwt.ParseWithClaims(in.Token, &FPClaim{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(v), nil
	})
	if err != nil {
		log.Printf("jwt parsing error,\n%v", err)
		return &pb.VerifyResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	claim := t.Claims.(*FPClaim)
	
	if !t.Valid || claim.Type != "AUTH" {
		return &pb.VerifyResponse{Valid: false}, nil
	}

	session := &model.UserSession{}

	collection := mClient.Database("farerpath").Collection("sessions")
	err = collection.FindOne(ctx, bson.D{{"authToken",in.GetToken()}}).Decode(session)
	if err != nil {
		log.Printf("DB find failed,\n%v", err)
	}

	return &pb.VerifyResponse{Valid: true, UserID: session.UserID}, nil
}

func (s *sessionService) DelSession(ctx context.Context, in *pb.Token) (*pb.Value, error) {
	collection := mClient.Database("farerpath").Collection("sessions")
	_, err := collection.DeleteOne(ctx, bson.D{{"authToken", in.Token}})
	if err != nil {
		log.Printf("DB deletion failed,\n%v", err)
		return &pb.Value{}, status.Error(codes.Internal, err.Error())
	}

	rClient.Del(in.GetToken())

	return &pb.Value{Value: "200"}, nil
}

func secureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal("Unable to generate random bytes")
	}
	return randomBytes
}

func secureRandomString(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

	// Compute bitMask
	availableCharLength := len(letters)
	if availableCharLength == 0 || availableCharLength > 256 {
		panic("availableCharBytes length must be greater than 0 and less than or equal to 256")
	}
	var bitLength byte
	var bitMask byte
	for bits := availableCharLength - 1; bits != 0; {
		bits = bits >> 1
		bitLength++
	}
	bitMask = 1<<bitLength - 1

	// Compute bufferSize
	bufferSize := length + length / 3

	// Create random string
	result := make([]byte, length)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			// Random byte buffer is empty, get a new one
			randomBytes = secureRandomBytes(bufferSize)
		}
		// Mask bytes to get an index into the character slice
		if idx := int(randomBytes[j%length] & bitMask); idx < availableCharLength {
			result[i] = letters[idx]
			i++
		}
	}

	return string(result)
}