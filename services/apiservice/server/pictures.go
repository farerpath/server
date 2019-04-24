package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/farerpath/server/common/util"
	"github.com/farerpath/server/model/consts"
	"github.com/farerpath/server/model/errors"
	"github.com/farerpath/server/model/model"

	"github.com/farerpath/randstr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"

	"github.com/jasonwinn/geocoder"

	"github.com/rwcarlsen/goexif/exif"

	albumService "github.com/farerpath/albumservice/proto"
)

const FILE_SERVICE_URL = "http://192.168.1.151/"

func UploadPicture(userId, pictureName string, publishRange uint, file *multipart.File, fileHeader *multipart.FileHeader) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	pictureId := randstr.GenerateRandomString(16)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileHeader.Filename)

	if err != nil {
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	_, err = io.Copy(part, *file)
	if err != nil {
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	_ = writer.WriteField("userId", userId)
	_ = writer.WriteField("pictureId", pictureId)

	err = writer.Close()
	if err != nil {
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	req, err := http.NewRequest(http.MethodPost, FILE_SERVICE_URL, body)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	if !util.IsSucced(int32(uploadResp.StatusCode)) {
		returnValue.StatusCode = uploadResp.StatusCode
		return
	}

	geocoder.SetAPIKey(consts.GOOGLE_REVGEOCODING_KEY)

	path := &albumService.Path{}

	metadata, err := exif.Decode(*file)
	if err == nil {
		// Cannot decode file.
		// But not critical.
		// PASS_THROUGH
		lat, lng, err := metadata.LatLong()
		if err == nil {
			// Cannot decode file.
			// But not critical.
			// PASS_THROUGH

			path.Location = &albumService.GPSData{Latitude: strconv.FormatFloat(lat, 'f', -1, 64), Longitude: strconv.FormatFloat(lng, 'f', -1, 64)}
		}

		addr, err := geocoder.ReverseGeocode(lat, lng)
		if err == nil {
			// Cannot get location.
			// But not critical.
			// PASS_THROUGH

			path.Country = addr.CountryCode
			path.City = addr.City
		}
	}

	makePicReq := &albumService.MakePictureRequest{
		// PictureId:     pictureId,// TODO: fix
		PictureName:   pictureName,
		Owner:         userId,
		CountNiceShot: 0,
		PublishRange:  uint32(publishRange),
		Archived:      false,
		Path:          path,
	}

	_, err = albumClient.MakePicture(context.Background(), makePicReq)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	return
}

func DownloadPicture(userId, albumId, pictureId string, fileName *string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	resp, err := albumClient.GetPicture(context.Background(), &albumService.GetPictureRequest{ReqUserID: userId, AlbumID: albumId, PictureID: pictureId})
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

	picResp, err := http.Get(FILE_SERVICE_URL + "?userId=" + userId + "&pictureId=" + pictureId)

	body, err := ioutil.ReadAll(picResp.Body)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	*fileName = resp.GetPictureName()

	returnValue.Value = body
	return
}

func ArchivePicture(userID, pictureID string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := albumClient.ArchivePicture(context.Background(), &albumService.ArchivePictureRequest{UserID: userID, PictureID: pictureID})
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

func RecoverPicture(userID, pictureID string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := albumClient.RecoverPicture(context.Background(), &albumService.RecoverPictureRequest{ReqUserID: userID, PictureID: pictureID})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
			return
		} else if st.Code() == codes.AlreadyExists {
			returnValue.StatusCode = http.StatusConflict
			return
		}

		log.Println(err)
		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	return
}

func DeletePicture(userId, pictureId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := albumClient.DestroyPicture(context.Background(), &albumService.DestroyPictureRequest{UserID: userId, PictureID: pictureId})
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

	client := &http.Client{}
	data := url.Values{"userId": {userId}, "pictureId": {pictureId}}

	req, err := http.NewRequest(http.MethodDelete, FILE_SERVICE_URL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	fileResp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	returnValue.StatusCode = fileResp.StatusCode
	return
}

func GetPictureList(userId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	resp, err := albumClient.GetPictureList(context.Background(), &albumService.GetPictureListRequest{ReqUserID: userId, DstUserID: userId, Archived: false})
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

	picList := []model.Picture{}

	for _, picture := range resp.GetPictureList() {
		path := model.Path{
			City:    picture.GetPath().GetCity(),
			Country: picture.GetPath().GetCountry(),
			Location: model.GPSData{
				Latitude:  picture.GetPath().GetLocation().GetLatitude(),
				Longitude: picture.GetPath().GetLocation().GetLongitude(),
				Altitude:  picture.GetPath().GetLocation().GetAltitude(),
			},
		}

		comments := []model.Comment{}

		for _, comment := range picture.Comments {
			comments = append(comments, model.Comment{
				Owner:     comment.GetOwner(),
				UserName:  comment.GetUserName(),
				Value:     comment.GetValue(),
				Time:      comment.GetTime(),
				//CommentID: comment.GetCommentId(), //TODO:fix
			})
		}

		// TODO:fix
		tmp := model.Picture{
			PictureName:   picture.GetPictureName(),
			//PictureID:     picture.GetPictureId(),
			Owner:         picture.GetOwner(),
			//TimeMeatadata: picture.GetTimeMetadata(),
			//PublishRange:  uint(picture.GetPublishRange()),
			//CountNiceShot: uint(picture.GetCountNiceShot()),
			Archived:      picture.GetArchived(),
			Path:          path,
			Comments:      comments,
		}

		picList = append(picList, tmp)
	}

	returnValue.Value = util.MakeReturnValueToJson(picList)

	return
}

func GetArchivedPictureList(userID string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()
	resp, err := albumClient.GetPictureList(context.Background(), &albumService.GetPictureListRequest{ReqUserID: userID, DstUserID: userID, Archived: true})
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

	picList := []model.Picture{}

	for _, picture := range resp.GetPictureList() {
		path := model.Path{
			City:    picture.GetPath().GetCity(),
			Country: picture.GetPath().GetCountry(),
			Location: model.GPSData{
				Latitude:  picture.GetPath().GetLocation().GetLatitude(),
				Longitude: picture.GetPath().GetLocation().GetLongitude(),
				Altitude:  picture.GetPath().GetLocation().GetAltitude(),
			},
		}

		comments := []model.Comment{}

		// TODO:fix
		for _, comment := range picture.Comments {
			comments = append(comments, model.Comment{
				Owner:     comment.GetOwner(),
				UserName:  comment.GetUserName(),
				Value:     comment.GetValue(),
				Time:      comment.GetTime(),
				//CommentID: comment.GetCommentId(),
			})
		}

		// TODO: fix
		tmp := model.Picture{
			PictureName:   picture.GetPictureName(),
			//PictureID:     picture.GetPictureId(),
			Owner:         picture.GetOwner(),
			//TimeMeatadata: picture.GetTimeMetadata(),
			//PublishRange:  uint(picture.GetPublishRange()),
			//CountNiceShot: uint(picture.GetCountNiceShot()),
			Archived:      picture.GetArchived(),
			Path:          path,
			Comments:      comments,
		}

		picList = append(picList, tmp)
	}

	returnValue.Value = util.MakeReturnValueToJson(picList)

	return
}
