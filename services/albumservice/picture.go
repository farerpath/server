package main

import (
	"context"
	"time"

	"github.com/farerpath/server/model/uoid"

	pb "github.com/farerpath/albumservice/proto"
	"github.com/farerpath/server/model/model"
	"go.mongodb.org/mongo-driver/bson"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"log"
)

// MakePicture func
// Make picture structure
// Add picture to default album
func (srv *albumService) MakePicture(ctx context.Context, req *pb.MakePictureRequest) (*pb.MakePictureReply, error) {
	path := model.Path{
		City:     req.GetPath().GetCity(),
		Country:  req.GetPath().GetCountry(),
		Location: model.GPSData(*req.GetPath().GetLocation()),
	}

	picture := &model.Picture{
		PictureID:     uoid.New(),
		PictureName:   req.GetPictureName(),
		Owner:         req.GetOwner(),
		TimeMetadata:  time.Unix(req.GetTimeMetadata(), 0),
		PublishRange:  req.GetPublishRange(),
		CountNiceShot: req.GetCountNiceShot(),
		Path:          path,
	}

	_, err := pictureCollection.InsertOne(ctx, picture)
	if err != nil {
		log.Printf("Error accured: MakePicture\n%v", err)
		return &pb.MakePictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.MakePictureReply{}, nil
}

// AddPicture func
// Add picture to specific album
func (srv *albumService) AddPicture(ctx context.Context, req *pb.AddPictureRequest) (*pb.AddPictureReply, error) {
	album := &model.Album{}
	err := albumCollection.FindOne(ctx, bson.D{{"albumID", req.GetAlbumID()}}).Decode(album)
	if err != nil {
		return &pb.AddPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	found := false
	for _, member := range album.Members {
		if member == req.GetReqUserID() {
			found = true
			break
		}
	}

	if req.GetReqUserID() == album.Owner {
		found = true
	}

	if !found {
		return &pb.AddPictureReply{}, status.Error(codes.Unauthenticated, "")
	}

	picture := &model.Picture{}
	err = pictureCollection.FindOne(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetReqUserID()}}).Decode(picture)
	if err != nil {
		return &pb.AddPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	for _, pic := range album.Pictures {
		if pic.String() == req.GetPictureID() {
			album.Pictures = append(album.Pictures, pic)
			break
		}
	}

	_, err = albumCollection.UpdateOne(ctx, bson.D{{"albumID", req.GetAlbumID()}}, album)
	if err != nil {
		return &pb.AddPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddPictureReply{}, nil
}

func (srv *albumService) GetPicture(ctx context.Context, req *pb.GetPictureRequest) (*pb.GetPictureReply, error) {
	album := &model.Album{}

	err := albumCollection.FindOne(ctx, bson.D{{"owner", req.GetReqUserID()}, {"albumID", req.GetAlbumID()}}).Decode(album)
	if err != nil {
		log.Printf("Error accured: GetPicture\n%v", err)
		return &pb.GetPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	picture := &model.Picture{}

	err = pictureCollection.FindOne(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetReqUserID()}}).Decode(picture)
	if err != nil {
		log.Printf("Error accured: GetPicture\n%v", err)
		return &pb.GetPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	l := pb.GPSData(picture.Path.Location)
	path := &pb.Path{
		City:     picture.Path.City,
		Country:  picture.Path.Country,
		Location: &l,
	}

	result := &pb.GetPictureReply{
		PictureName:   picture.PictureName,
		PictureID:     picture.PictureID.String(),
		Owner:         picture.Owner,
		TimeMetadata:  picture.TimeMetadata.Unix(),
		PublishRange:  picture.PublishRange,
		CountNiceShot: picture.CountNiceShot,
		Archived:      picture.Archived,
		Path:          path,
	}

	for _, pic := range album.Pictures {
		if pic.String() == req.GetPictureID() {
			return result, nil
		}
	}

	return &pb.GetPictureReply{}, status.Error(codes.NotFound, "picture not found from album")
}

// DelPicture func
// Remove Picture from album
func (srv *albumService) DelPicture(ctx context.Context, req *pb.DelPictureRequest) (*pb.DelPictureReply, error) {
	n, err := pictureCollection.CountDocuments(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetReqUserID()}})
	if n == 0 {
		return &pb.DelPictureReply{}, status.Error(codes.NotFound, "")
	}
	if err != nil {
		return &pb.DelPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	album := &model.Album{}
	err = albumCollection.FindOne(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetReqUserID()}}).Decode(album)
	if err != nil {
		return &pb.DelPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	found := false
	for _, member := range album.Members {
		if member == req.GetReqUserID() {
			found = true
			break
		}
	}

	if !found {
		return &pb.DelPictureReply{}, status.Error(codes.Unauthenticated, "")
	}

	found = false
	for idx, pic := range album.Pictures {
		if pic.String() == req.GetPictureID() {
			album.Pictures = append(album.Pictures[:idx], album.Pictures[idx+1:]...)
			found = true
			break
		}
	}

	if !found {
		return &pb.DelPictureReply{}, status.Error(codes.NotFound, "")
	}

	_, err = albumCollection.UpdateOne(ctx, bson.D{{"albumID", req.GetAlbumID()}}, album)
	if err != nil {
		return &pb.DelPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.DelPictureReply{}, nil
}

func (srv *albumService) ArchivePicture(ctx context.Context, req *pb.ArchivePictureRequest) (*pb.ArchivePictureReply, error) {
	picture := &model.Picture{}

	err := pictureCollection.FindOne(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetUserID()}}).Decode(picture)

	if picture.Archived {
		return &pb.ArchivePictureReply{}, status.Error(codes.AlreadyExists, "")
	}

	_, err = pictureCollection.UpdateOne(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetUserID()}}, bson.D{{"$set", bson.D{{"archived", true}}}})
	if err != nil {
		return &pb.ArchivePictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.ArchivePictureReply{}, nil
}

func (srv *albumService) DestroyPicture(ctx context.Context, req *pb.DestroyPictureRequest) (*pb.DestroyPictureReply, error) {
	_, err := pictureCollection.DeleteOne(ctx, bson.D{{"pictureID", req.GetPictureID()}, {"owner", req.GetUserID()}})
	if err != nil {
		return &pb.DestroyPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.DestroyPictureReply{}, nil
}

func (srv *albumService) GetPictureList(ctx context.Context, req *pb.GetPictureListRequest) (*pb.GetPictureListReply, error) {
	pictures := []model.Picture{}
	cur, err := pictureCollection.Find(ctx, bson.D{{"owner", req.GetReqUserID()}, {"archived", req.GetArchived()}})
	if err != nil {
		return &pb.GetPictureListReply{}, status.Error(codes.Internal, err.Error())
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		picture := model.Picture{}
		cur.Decode(picture)
		pictures = append(pictures, picture)
	}

	resp := &pb.GetPictureListReply{}

	for _, picture := range pictures {
		path := &pb.Path{
			City:    picture.Path.City,
			Country: picture.Path.Country,
			Location: &pb.GPSData{
				Latitude:  picture.Path.Location.Latitude,
				Longitude: picture.Path.Location.Longitude,
				Altitude:  picture.Path.Location.Altitude,
			},
		}

		comments := []*pb.Comment{}

		for _, comment := range picture.Comments {
			comments = append(comments, &pb.Comment{
				Owner:     comment.Owner,
				UserName:  comment.UserName,
				Value:     comment.Value,
				Time:      comment.Time,
				CommentID: comment.CommentID.String(),
			})
		}

		tmp := &pb.Picture{
			PictureName:   picture.PictureName,
			PictureID:     picture.PictureID.String(),
			Owner:         picture.Owner,
			TimeMetadata:  picture.TimeMetadata.Unix(),
			PublishRange:  uint32(picture.PublishRange),
			CountNiceShot: uint32(picture.CountNiceShot),
			Archived:      picture.Archived,
			Path:          path,
			Comments:      comments,
		}

		resp.PictureList = append(resp.PictureList, tmp)
	}

	return resp, nil
}

func (srv *albumService) RenamePicture(ctx context.Context, req *pb.RenamePictureRequest) (*pb.RenamePictureReply, error) {
	_, err := pictureCollection.UpdateOne(ctx, bson.D{{"owner", req.GetReqUserID()}, {"pictureID", req.GetPictureID()}}, bson.D{{"$set", bson.D{{"pictureName", req.GetPictureName()}}}})
	if err != nil {
		return &pb.RenamePictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.RenamePictureReply{}, nil
}

func (srv *albumService) RecoverPicture(ctx context.Context, req *pb.RecoverPictureRequest) (*pb.RecoverPictureReply, error) {
	picture := &model.Picture{}
	err := pictureCollection.FindOne(ctx, bson.D{{"userID", req.GetReqUserID()}, {"pictureID", req.GetPictureID()}}).Decode(picture)
	if err != nil {
		return &pb.RecoverPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	if !picture.Archived {
		return &pb.RecoverPictureReply{}, status.Error(codes.AlreadyExists, "")
	}

	_, err = pictureCollection.UpdateOne(ctx, bson.D{{"userID", req.GetReqUserID()}, {"picutreID", req.GetPictureID()}}, bson.D{{"$set", bson.D{{"archived", false}}}})
	if err != nil {
		return &pb.RecoverPictureReply{}, status.Error(codes.Internal, err.Error())
	}

	return &pb.RecoverPictureReply{}, nil
}

func (srv *albumService) AddComment(ctx context.Context, req *pb.AddCommentRequest) (*pb.AddCommentReply, error) {
	return &pb.AddCommentReply{}, nil
}

func (srv *albumService) DelComment(ctx context.Context, req *pb.DelCommentRequest) (*pb.DelCommentReply, error) {
	return &pb.DelCommentReply{}, nil
}

func (srv *albumService) AddNiceShot(ctx context.Context, req *pb.AddNiceShotRequest) (*pb.AddNiceShotReply, error) {
	return &pb.AddNiceShotReply{}, nil
}

func (srv *albumService) SubNiceShot(ctx context.Context, req *pb.SubNiceShotRequest) (*pb.SubNiceShotReply, error) {
	return &pb.SubNiceShotReply{}, nil
}
