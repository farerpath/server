package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	pb "github.com/farerpath/albumservice/proto"
	"github.com/farerpath/server/model/consts"
	"github.com/farerpath/server/model/model"
	"github.com/farerpath/server/model/uoid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"log"
)

// MakeAlbumList func
// MakeAlbumList to user, initialize user album list
// Make default album
func (srv *albumService) MakeAlbumList(ctx context.Context, req *pb.MakeAlbumListRequest) (*pb.MakeAlbumListReply, error) {
	defaultAlbumList := &model.AlbumList{
		UserID: req.ReqUserID,
	}

	_, err := albumListCollection.InsertOne(ctx, defaultAlbumList)
	if err != nil {
		log.Printf("Error accured: MakeAlbumList\n%v", err)
		return &pb.MakeAlbumListReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.MakeAlbumListReply{}, nil
}

// TODO: Aggregate
func (srv *albumService) GetAlbumList(ctx context.Context, req *pb.GetAlbumListRequest) (*pb.GetAlbumListReply, error) {
	fullAlbumList := &model.AlbumListWithAlbumWithPicture{}
	albumList := &model.AlbumList{}

	err := albumListCollection.FindOne(ctx, bson.D{{"_id", req.GetDstUserID()}}).Decode(albumList)
	if err != nil {
		log.Printf("Error accured: GetAlbumList\n%v", err)
		return &pb.GetAlbumListReply{}, status.Error(codes.Internal, err.Error())
	}

	if len(albumList.Albums) < 1 {
		return &pb.GetAlbumListReply{UserID:albumList.UserID, AlbumList:[]*pb.AlbumNode{}}, nil
	}

	cur, err := albumListCollection.Aggregate(ctx, mongo.Pipeline{
		{{"$lookup", bson.D{{"from", "albums"}, {"localField", "albums"}, {"foreignField", "_id"}, {"as", "albums"}}}},
		{{"$unwind", bson.D{{"path", "$albums"}, {"preserveNullAndEmptyArrays", true}}}},
		{{"$lookup", bson.D{{"from", "picture"}, {"localField", "albums.pictures"}, {"foreignField", "_id"}, {"as", "albums.pictures"}}}},
		{{"$group", bson.D{{"_id", "$_id"}, {"albums", bson.D{{"$push", "$albums"}}}}}},
	}, options.Aggregate())

	cur.Next(ctx)

	err = cur.Decode(&fullAlbumList)
	if err != nil {
		log.Printf("Error accured: GetAlbumList\n%v", err)
		return &pb.GetAlbumListReply{}, status.Error(codes.Internal, err.Error())
	}

	result := &pb.GetAlbumListReply{
		UserID: albumList.UserID,
	}


	for _, album := range fullAlbumList.Albums {
		l := pb.GPSData(album.TravelPath.Location)
		p := &pb.Path{
			Country:  album.TravelPath.Country,
			City:     album.TravelPath.City,
			Location: &l,
		}
		a := &pb.AlbumNode{
			//AlbumID:    album.AlbumID,
			AlbumName:  album.AlbumName,
			Owner:      album.Owner,
			BeginTime:  album.BeginTime.Unix(),
			EndTime:    album.EndTime.Unix(),
			TravelPath: p,
		}

		result.AlbumList = append(result.AlbumList, a)
	}

	//log.Printf("Result: %v\n", result.AlbumList[0].Owner)
	// NOT YET IMPLEMENTED
	// After aggregate
	/*
	if result.GetUserID() != req.GetReqUserID() {
		reply := &pb.GetAlbumListReply{UserID: result.GetUserID()}
		
			for _, node := range result.AlbumList {
				if node.IsPublic {
					reply.AlbumList = append(reply.AlbumList, node)
				}
			}
		

		return reply, nil
	}*/

	return result, nil
}

func (srv *albumService) DelAlbumList(ctx context.Context, req *pb.DelAlbumListRequest) (*pb.DelAlbumListReply, error) {
	_, err := albumListCollection.DeleteOne(ctx, bson.D{{"_id", req.GetReqUserID()}})
	if err != nil {
		log.Printf("Error accured: DelAlbumList\n%v", err)
		return &pb.DelAlbumListReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.DelAlbumListReply{}, nil
}

func (srv *albumService) MakeAlbum(ctx context.Context, req *pb.MakeAlbumRequest) (*pb.MakeAlbumReply, error) {
	album := &model.Album{
		AlbumID: 		uoid.New(),
		AlbumName:    	req.GetAlbumName(),
		Owner:        	req.GetOwner(),
		BeginTime:    	time.Unix(req.GetBeginTime(), 0),
		EndTime:      	time.Unix(req.GetEndTime(), 0),
		Members:      	req.GetMembers(),
		PublishRange: 	req.GetPublishRange(),
	}

	_, err := albumCollection.InsertOne(ctx, album)
	if err != nil {
		log.Printf("Error accured: MakeAlbum\n%v", err)
		return &pb.MakeAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	isPublic := false
	if req.PublishRange == consts.PUBLIC {
		isPublic = true
	}

	members := []string{album.Owner}
	members = append(members, album.Members...)

	result := addAlbumToAlbumList(members, album.AlbumID, isPublic)
	if result != 200 {
		return &pb.MakeAlbumReply{}, status.Error(codes.Internal, string(result))
	}

	return &pb.MakeAlbumReply{}, nil
}

func (srv *albumService) GetAlbum(ctx context.Context, req *pb.GetAlbumRequest) (*pb.GetAlbumReply, error) {
	album := &model.Album{}

	err := albumCollection.FindOne(ctx, bson.D{{"owner", req.GetReqUserID()}, {"albumID", req.GetAlbumID()}}).Decode(album)
	if err != nil {
		log.Printf("Error accured: GetAlbum\n%v", err)
		return &pb.GetAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	var pictures []string

	for _, pic := range album.Pictures {
		pictures = append(pictures, pic.String())
	}

	result := &pb.GetAlbumReply{
		AlbumName:    album.AlbumName,
		Owner:        album.Owner,
		BeginTime:    album.BeginTime.Unix(),
		EndTime:      album.EndTime.Unix(),
		Members:      album.Members,
		Pictures:     pictures,
		PublishRange: album.PublishRange,
	}


	return result, nil
}

func (srv *albumService) GetPublicAlbum(ctx context.Context, req *pb.GetPublicAlbumRequest) (*pb.GetPublicAlbumReply, error) {
	album := &model.Album{}

	err := albumCollection.FindOne(ctx, bson.D{{"owner", req.DstUserID}, {"albumID", req.GetAlbumID()}}).Decode(album)
	if err != nil {
		log.Printf("Error accured: GetPublicAlbum\n%v", err)
		return &pb.GetPublicAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	var pictures []string

	for _, pic := range album.Pictures {
		pictures = append(pictures, pic.String())
	}

	result := &pb.GetPublicAlbumReply{
		AlbumID:      album.AlbumID.String(),
		AlbumName:    album.AlbumName,
		Owner:        album.Owner,
		BeginTime:    album.BeginTime.Unix(),
		EndTime:      album.EndTime.Unix(),
		Members:      album.Members,
		Pictures:     pictures,
		PublishRange: album.PublishRange,
	}

	if result.GetPublishRange() == consts.PUBLIC {
		return result, nil
	}

	return &pb.GetPublicAlbumReply{}, status.Error(codes.NotFound, "")
}

func (srv *albumService) DelAlbum(ctx context.Context, req *pb.DelAlbumRequest) (*pb.DelAlbumReply, error) {
	_, err := albumCollection.DeleteOne(ctx, bson.D{{"owner", req.GetReqUserID()}, {"albumID", req.GetAlbumID()}})
	if err != nil {
		log.Printf("Error accured: DelAlbum\n%v", err)
		return &pb.DelAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.DelAlbumReply{}, nil
}

func (srv *albumService) RenameAlbum(ctx context.Context, req *pb.RenameAlbumRequest) (*pb.RenameAlbumReply, error) {
	_, err := albumCollection.UpdateOne(ctx, bson.D{{"userID", req.GetUserID()}, {"albumID", req.GetAlbumID()}}, bson.D{{"$set", bson.D{{"albumName", req.AlbumName}}}})
	if err != nil {
		log.Printf("Error accured: RenameAlbum\n%v", err)
		return &pb.RenameAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.RenameAlbumReply{}, nil
}

func (srv *albumService) EndAlbum(ctx context.Context, req *pb.EndAlbumRequest) (*pb.EndAlbumReply, error) {
	_, err := albumCollection.UpdateOne(ctx, bson.D{{"albumID", req.GetAlbumID()}}, bson.D{{"$set", bson.D{{"endTime", time.Now().UTC()}}}})

	if err != nil {
		log.Printf("Error accured: EndAlbum\n%v", err)
		return &pb.EndAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.EndAlbumReply{}, nil
}

func (srv *albumService) PublishAlbum(ctx context.Context, req *pb.PublishAlbumRequest) (*pb.PublishAlbumReply, error) {
	_, err := albumCollection.UpdateOne(ctx, bson.D{{"albumID", req.GetAlbumID()}}, bson.D{{"$set", bson.D{{"publishRange", consts.PUBLIC}}}})
	if err != nil {
		log.Printf("Error accured: PublishAlbum\n%v", err)
		return &pb.PublishAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublishAlbumReply{}, nil
}

func (srv *albumService) AddMember(ctx context.Context, req *pb.AddMemberRequest) (*pb.AddMemberReply, error) {
	album := &model.Album{}

	err := albumCollection.FindOne(ctx, bson.D{{"albumID", req.AlbumID}}).Decode(album)
	if err != nil {
		return &pb.AddMemberReply{}, status.Error(codes.Internal, err.Error())
	}

	isPublic := false
	if album.PublishRange == consts.PUBLIC {
		isPublic = true
	}

	result := addAlbumToAlbumList(req.GetMemberID(), uoid.FromString(req.AlbumID), isPublic)
	if result != 200 {
		return &pb.AddMemberReply{}, nil
	}

	return &pb.AddMemberReply{}, status.Error(codes.Internal, string(result))
}

func (srv *albumService) DelMember(ctx context.Context, req *pb.DelMemberRequest) (*pb.DelMemberRelpy, error) {
	album := &model.Album{}

	err := albumCollection.FindOne(ctx, bson.D{{"albumID", req.AlbumID}}).Decode(album)
	if err != nil {
		return &pb.DelMemberRelpy{}, status.Error(codes.Internal, err.Error())
	}

	result := deleteAlbumFromAlbumList(req.GetMemberID(), uoid.FromString(req.AlbumID))
	if result != 200 {
		return &pb.DelMemberRelpy{}, nil
	}

	return &pb.DelMemberRelpy{}, status.Error(codes.Internal, string(result))
}

func (srv *albumService) ArchiveAlbum(ctx context.Context, req *pb.ArchiveAlbumRequest) (*pb.ArchiveAlbumReply, error) {
	album := &model.Album{}
	err := albumCollection.FindOne(ctx, bson.D{{"albumID", req.AlbumID}}).Decode(album)
	if album.Archived {
		return &pb.ArchiveAlbumReply{}, status.Error(codes.AlreadyExists, "")
	}

	_, err = albumCollection.UpdateOne(ctx, bson.D{{"albumID", req.GetAlbumID()}}, bson.D{{"$set", bson.D{{"archived", true}}}})
	if err != nil {
		log.Printf("Error accured: DelAlbum\n%v", err)
		return &pb.ArchiveAlbumReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.ArchiveAlbumReply{}, nil
}

// TODO: update member's albumlist
// TODO: implement transaction
func addAlbumToAlbumList(userID []string, albumID uoid.UOID, isPublic bool) int {
	ctx := context.Background()
	for _, user := range userID {
		albumList := &model.AlbumList{}
		err := albumListCollection.FindOne(ctx, bson.D{{"_id", user}}).Decode(albumList)
		if err != nil {
			log.Println(err)
			continue
		}

		if len(albumList.Albums) < 1 {
			albumList.Albums = append(albumList.Albums, albumID)
		} else {
			// check if already exists
			for _, album := range albumList.Albums {
				if album == albumID {
					continue
				}
				albumList.Albums = append(albumList.Albums, albumID)
			}
		}

		log.Println(albumList.Albums)

		_, err = albumListCollection.ReplaceOne(ctx, bson.D{{"_id", user}}, albumList)
		if err != nil {
			log.Printf("Replace fail %v\n", err)
			continue
		}
	}

	return 200
}

// TODO: update member's albumlist
func deleteAlbumFromAlbumList(userID []string, albumID uoid.UOID) int {
	ctx := context.Background()

	for _, user := range userID {
		albumList := &model.AlbumList{}

		err := albumListCollection.FindOne(ctx, bson.D{{"_id", user}}).Decode(albumList)
		if err != nil {
			continue
		}

		for idx, album := range albumList.Albums {
			if album == albumID {
				albumList.Albums = append(albumList.Albums[:idx], albumList.Albums[idx+1:]...)
			}
		}

		_, err = albumListCollection.UpdateOne(ctx, bson.D{{"_id", user}}, albumList)
		if err != nil {
			continue
		}
	}

	return 200
}
