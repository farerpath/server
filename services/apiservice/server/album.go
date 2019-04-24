package server

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/farerpath/server/common/util"
	"github.com/farerpath/server/model/errors"
	"github.com/farerpath/server/model/model"

	"golang.org/x/net/context"

	"strings"

	albumService "github.com/farerpath/albumservice/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetAlbumList(sessionUserId, userId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	resp, err := albumClient.GetAlbumList(context.Background(), &albumService.GetAlbumListRequest{ReqUserID: sessionUserId, DstUserID: userId})
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	/*
		albumList := make(map[string]interface{})
		albumList["userID"] = resp.GetUserId()
		albumList["albums"] = resp.AlbumList*/

	returnValue.Value = util.MakeReturnValueToJson(resp)

	return
}

func GetAlbum(sessionUserId, userId, albumId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	if sessionUserId != userId {
		resp, err := albumClient.GetPublicAlbum(context.Background(), &albumService.GetPublicAlbumRequest{DstUserID: userId, AlbumID: albumId})
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.NotFound {
				returnValue.StatusCode = http.StatusNotFound
				return
			}
			returnValue.StatusCode = http.StatusInternalServerError
			return
		}

		path := model.Path{
			Country:  resp.GetTravelPath().GetCountry(),
			City:     resp.GetTravelPath().GetCity(),
			Location: model.GPSData(*resp.GetTravelPath().GetLocation()),
		}

		album := &model.Album{
			AlbumName:    resp.GetAlbumName(),
			Owner:        resp.GetOwner(),
			BeginTime:    time.Unix(resp.BeginTime, 0),
			EndTime:      time.Unix(resp.GetEndTime(), 0),
			TravelPath:   path,
			PublishRange: resp.GetPublishRange(),
			Members:      resp.GetMembers(),
			//Pictures:     resp.GetPictures().Hex(), //TODO: fix
		}

		returnValue.Value = util.MakeReturnValueToJson(album)
		return
	}
	resp, err := albumClient.GetAlbum(context.Background(), &albumService.GetAlbumRequest{ReqUserID: userId, AlbumID: albumId})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
			return
		}
		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	path := model.Path{
		Country:  resp.GetTravelPath().GetCountry(),
		City:     resp.GetTravelPath().GetCity(),
		Location: model.GPSData(*resp.GetTravelPath().GetLocation()),
	}

	album := &model.Album{
		//AlbumID:      resp.GetAlbumId(), //TODO:fix
		AlbumName:    resp.GetAlbumName(),
		Owner:        resp.GetOwner(),
		BeginTime:    time.Unix(resp.GetBeginTime(), 0),
		EndTime:      time.Unix(resp.GetEndTime(), 0),
		TravelPath:   path,
		PublishRange: resp.GetPublishRange(),
		Members:      resp.GetMembers(),
		//Pictures:     resp.GetPictures(),// TODO:fix
	}

	returnValue.Value = util.MakeReturnValueToJson(album)
	return
}

func MakeAlbum(sessionUserId, userId, albumName, beginTime, members string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	if sessionUserId != userId {
		returnValue.StatusCode = http.StatusForbidden
		return
	}

	members = util.DelSpaces(members)

	memberList := strings.Split(members, ",")

	beginTimeParsed, err := strconv.ParseInt(beginTime, 0, 64)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = http.StatusBadRequest
		return
	}

	makeAlbumReq := &albumService.MakeAlbumRequest{
		Owner:     userId,
		AlbumName: albumName,
		BeginTime: beginTimeParsed,
		Members:   memberList,
	}

	_, err = albumClient.MakeAlbum(context.Background(), makeAlbumReq)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	returnValue.StatusCode = http.StatusOK

	return
}

func UpdateAlbum(userId, albumId string, newAlbumName, newBeginTime, newMembers string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	if len(newAlbumName) != 0 {
		_, err := albumClient.RenameAlbum(context.Background(), &albumService.RenameAlbumRequest{UserID: userId, AlbumID: albumId, AlbumName: newAlbumName})
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.NotFound {
				returnValue.StatusCode = http.StatusNotFound
				return
			}
			returnValue.StatusCode = http.StatusInternalServerError
			return
		}
	}

	members := strings.Split(util.DelSpaces(newMembers), ",")

	if len(newMembers) != 0 {
		_, err := albumClient.AddMember(context.Background(), &albumService.AddMemberRequest{ReqUserID: userId, AlbumID: albumId, MemberID: members})
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.NotFound {
				returnValue.StatusCode = http.StatusNotFound
				return
			}
			returnValue.StatusCode = http.StatusInternalServerError
			return
		}
	}

	if len(newBeginTime) != 0 {
		// NEED TO BE ADDED
	}

	return
}

func DeleteAlbum(userId, albumId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := albumClient.DelAlbum(context.Background(), &albumService.DelAlbumRequest{ReqUserID: userId, AlbumID: albumId})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
			return
		}

		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	return
}

func AddPictureToAlbum(userId, pictureId, albumId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := albumClient.AddPicture(context.Background(), &albumService.AddPictureRequest{ReqUserID: userId, PictureID: pictureId, AlbumID: albumId})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
			return
		}

		log.Println(err)

		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	return
}

func RemovePictureFromAlbum(userId, pictureId, albumId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := albumClient.DelPicture(context.Background(), &albumService.DelPictureRequest{ReqUserID: userId, PictureID: pictureId, AlbumID: albumId})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
			return
		}

		log.Println(err)

		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	return
}
