package main

import (
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"

	pb "github.com/farerpath/authservice/proto"

	"github.com/farerpath/server/model/model"

	psession "github.com/farerpath/sessionservice/proto"

	"net"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const port = ":17080"

type authServer struct{}

var (
	col *mongo.Collection
)

const (
	USERDB = "accounts"
)

var (
	DBADDR = "mongodb://maindb-service:27017"
)

const (
	SALT = ""
)

type query map[string]interface{}

const SESSIONSERVICEADDR = "sessionservice-service:17080"

var sclient psession.SessionClient

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

	log.Println("DB Connection succeed")

	col = client.Database("farerpath").Collection(USERDB)

	conn, err := grpc.Dial(SESSIONSERVICEADDR, grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connetc grpc Session Service %v", err)
	}

	sclient = psession.NewSessionClient(conn)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Failed to listen tcp %v", err)
	}

	g := grpc.NewServer()

	pb.RegisterAuthServer(g, &authServer{})

	reflection.Register(g)

	err = g.Serve(listener)
	if err != nil {
		log.Printf("Failed to start grpc server %v", err)
	}
}

func (a *authServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	resp := &pb.LoginReply{}

	result := &model.Account{}

	err := col.FindOne(ctx, bson.D{{"_id", req.LoginUserID}}).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return resp, status.Error(codes.NotFound, "id not found")
		}
		log.Printf("Error accured: Login(FindUser)\n%v", err)
		return resp, status.Error(codes.Unknown, err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(SALT+req.GetLoginUserID()+req.GetLoginPassword()))
	if err != nil {
		log.Printf("Failed to validate password. May password not match or data damaged \n%v", err)
		return resp, status.Error(codes.Unauthenticated, "401")
	}

	// Session
	// Session Service will return JWT Token
	// return THAT JWT Token
	token, err := sclient.NewSession(context.Background(), &psession.SessionRequest{
		UserID: req.GetLoginUserID(), LoginDeviceType: req.GetLoginDeviceType(), SessionDurationTime: int32(result.SessionDuration), LoginIP: req.GetLoginIP()})

	if err != nil {
		log.Printf("Failed to get token, service returned error \n%v", err)
		return resp, err
	}

	resp.UserID = result.UserID
	resp.UserName = result.UserName
	resp.Country = result.Country
	resp.RefreshToken = token.RefreshToken
	resp.AuthToken = token.AuthToken

	return resp, nil
}

func (a *authServer) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutReply, error) {
	// Session
	// Verify JWT Token in Session Service
	// if verified, Send DELETE request to Session Service
	v, err := sclient.VerifyToken(context.Background(), &psession.Token{Token: in.GetToken()})
	if err != nil {
		log.Printf("Failed to verify token. service returned error.\n%v", err)
		return nil, err
	}

	if v.Valid {
		verify, err := sclient.DelSession(context.Background(), &psession.Token{Token: in.GetToken()})
		if err != nil {
			return nil, err
		}

		if verify.Value == "200" {
			return &pb.LogoutReply{}, nil
		}
		return &pb.LogoutReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.LogoutReply{}, status.Error(codes.Unauthenticated, "")
}

func (a *authServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	resp := &pb.RegisterReply{}
	// Existance check

	n, err := col.CountDocuments(ctx, bson.D{{"_id", req.UserID}})
	if err != nil {
		log.Printf("Error accured: Register\n%v", err)
		return resp, status.Error(codes.Unknown, err.Error())
	}
	if n != 0 {
		return resp, status.Error(codes.AlreadyExists, "UserID already exists")
	}

	n, err = col.CountDocuments(ctx, bson.D{{"email", req.Email}})
	if err != nil {
		log.Printf("Error accured: Register\n%v", err)
		return resp, status.Error(codes.Unknown, err.Error())
	}
	if n != 0 {
		return resp, status.Error(codes.AlreadyExists, "Email already exists")
	}

	// Email confirm

	// If confirmed
	// Hash Password and Send to DB Service
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(SALT+req.GetUserID()+req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("password hashing failed\n %v", err)
		return resp, status.Error(codes.Internal, err.Error())
	}

	user := &model.Account{
		UserID:           req.GetUserID(),
		UserName:         "",
		Password:         string(hashedPassword),
		Email:            req.GetEmail(),
		FirstName:        "",
		LastName:         "",
		Country:          "",
		ProfilePhotoPath: "",
		Birthday:         time.Time{},
		SessionDuration:  3,
		IsBirthdayPublic: false,
		IsCountryPublic:  false,
		IsProfilePublic:  false,
	}

	_, err = col.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error accured: Register\n%v", err)
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (a *authServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	user := &model.Account{}

	err := col.FindOne(ctx, bson.D{{"_id", req.UserID}}).Decode(user)
	if err != nil {
		log.Printf("Error accured: Login(GetUser)\n%v", err)
		return &pb.GetUserReply{}, status.Error(codes.Unknown, err.Error())
	}

	reply := &pb.GetUserReply{
		Email:            user.Email,
		UserID:           user.UserID,
		UserName:         user.UserName,
		Country:          user.Country,
		Birthday:         user.Birthday.Unix(),
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		ProfilePhotoPath: user.ProfilePhotoPath,
		SessionDuration:  user.SessionDuration,
		IsProfilePublic:  user.IsProfilePublic,
		IsCountryPublic:  user.IsCountryPublic,
		IsBirthdayPublic: user.IsBirthdayPublic,
	}

	return reply, nil
}
