package route

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/farerpath/server/services/apiservice/server"

	"fmt"

	"github.com/farerpath/server/model/errors"
	"github.com/farerpath/server/model/model"

	sessionService "github.com/farerpath/sessionservice/proto"

	"github.com/mssola/user_agent"
	"golang.org/x/net/context"

	"github.com/farerpath/server/common/util"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var sessionClient sessionService.SessionClient

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong"))
}

func InitGrpcConn(conn *grpc.ClientConn) {
	sessionClient = sessionService.NewSessionClient(conn)
}

func Login(w http.ResponseWriter, r *http.Request) {
	body, err := util.UnmarshalBody(&r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	loginUserId := body["userID"].(string)
	loginPassword := body["password"].(string)
	token := r.Header.Get("X-Farerpath-Token")

	if len(token) > 10 {
		resp, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resp.GetValid() {
			result := server.TokenLogin(token, resp.GetUserID())
			w.WriteHeader(result.StatusCode)
			w.Write(result.Value)
		}
	}

	if (len(loginUserId) < 4 || len(loginPassword) < 8) && len(token) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	loginDeviceType := 0;
	if _, exists := body["loginDeviceType"]; exists {
		var err error
		loginDeviceType, err = strconv.Atoi(body["loginDeviceType"].(string))
		if err != nil {
			loginDeviceType = 0
		}
	}
	
	ua := user_agent.New(r.UserAgent())
	if ua.Bot() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := server.Login(loginUserId, loginPassword, loginDeviceType, util.GetReqIP(r.Header.Get("X-Forwarded-For")))

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Farerpath-Token")

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	result := server.Logout(token)

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := util.UnmarshalBody(&r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := server.Register(body["userID"].(string), body["email"].(string), body["password"].(string))

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

// TODO: Complete
func UserHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Farerpath-Token")
	vars := mux.Vars(r)
	result := &model.ReturnValue{}

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if r.Method == http.MethodGet {
		result = server.GetUserInformation(verify.UserID, vars["userId"])
	} else {
		isCountryPublic, err := strconv.ParseBool(r.FormValue("isCountryPublic"))
		if err != nil {
			w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
			return
		}

		isBirthdayPublic, err := strconv.ParseBool(r.FormValue("isBirthdayPublic"))
		if err != nil {
			w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
			return
		}

		isProfilePublic, err := strconv.ParseBool(r.FormValue("isProfilePublic"))
		if err != nil {
			w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
			return
		}

		result = server.UpdateUserProfile(verify.UserID, vars["userId"], r.FormValue("userName"), r.FormValue("email"), r.FormValue("profilePhotoPath"), r.FormValue("country"), r.FormValue("birthday"), isCountryPublic, isBirthdayPublic, isProfilePublic)
	}

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)

}

// TODO: Complete
func UserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Farerpath-Token")
	vars := mux.Vars(r)

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	result := server.UpdateUserPassword(verify.UserID, vars["userId"], r.FormValue("oldPassword"), r.FormValue("newPassword"))

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func AlbumHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	result := &model.ReturnValue{}

	token := r.Header.Get("X-Farerpath-Token")
	userId := vars["userId"]

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to verify session. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if r.Method == http.MethodGet {
		// Get Album list
		result = server.GetAlbumList(verify.UserID, userId)
	} else if r.Method == http.MethodPost {
		// Make Album
		body, err := util.UnmarshalBody(&r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		i := func() int {
			if _, exists := body["albumName"]; !exists {
				return -1
			}
			if _, exists := body["beginTime"]; !exists {
				return -1
			}
			if _, exists := body["members"]; !exists {
				return -1
			}
			return 0
		}()

		if i != 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Arguments are not filled"))
			return 
		}

		result = server.MakeAlbum(verify.UserID, userId, body["albumName"].(string), body["beginTime"].(string), body["members"].(string))
	}

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func UserAlbumHandler(w http.ResponseWriter, r *http.Request) {
	var result *model.ReturnValue
	vars := mux.Vars(r)

	token := r.Header.Get("X-Farerpath-Token")
	userID := vars["userId"]
	albumID := vars["albumId"]

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if r.Method == http.MethodGet {
		result = server.GetAlbum(verify.UserID, userID, albumID)
	} else if r.Method == http.MethodDelete {
		if verify.UserID != userID {
			w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
			return
		}

		result = server.DeleteAlbum(userID, albumID)
	} else if r.Method == http.MethodPatch {
		if verify.UserID != userID {
			w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
			return
		}

		body, err := util.UnmarshalBody(&r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result = server.UpdateAlbum(userID, albumID, body["newAlbumName"].(string), body["newBeginTime"].(string), body["newMembers"].(string))
	} else {
		w.WriteHeader(errors.STATUS_METHOD_NOT_ALLOWED)
		return
	}

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func AlbumPictureHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var result *model.ReturnValue

	token := r.Header.Get("X-Farerpath-Token")
	userID := vars["userId"]
	albumID := vars["albumId"]
	pictureID := vars["pictureId"]

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if verify.UserID != userID {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if r.Method == http.MethodPost {
		result = server.AddPictureToAlbum(userID, pictureID, albumID)
	} else if r.Method == http.MethodDelete {
		result = server.RemovePictureFromAlbum(userID, pictureID, albumID)
	} else if r.Method == http.MethodGet {
		var fileName string
		result = server.DownloadPicture(userID, albumID, pictureID, &fileName)
		http.ServeContent(w, r, fileName, time.Now(), bytes.NewReader(result.Value))
		return
	}
	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func PictureListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	result := &model.ReturnValue{}

	token := r.Header.Get("X-Farerpath-Token")
	userId := vars["userId"]

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if verify.UserID != userId {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if r.Method == http.MethodPost {
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
			return
		}
		publishRange, err := strconv.Atoi(r.FormValue("publishRange"))
		if err != nil {
			w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
			return
		}

		result = server.UploadPicture(userId, r.FormValue("pictureName"), uint(publishRange), &file, fileHeader)

	} else if r.Method == http.MethodGet {
		if r.FormValue("archived") == "true" {
			result = server.GetArchivedPictureList(userId)
		} else {
			result = server.GetPictureList(userId)
		}
	}

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

func PictureHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	result := &model.ReturnValue{}

	token := r.Header.Get("X-Farerpath-Token")
	userId := vars["userId"]
	pictureId := vars["pictureId"]

	verify, err := sessionClient.VerifyToken(context.Background(), &sessionService.Token{Token: token})
	if err != nil {
		fmt.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		w.WriteHeader(errors.STATUS_INTERNAL_ERROR)
		return
	}

	if !verify.Valid {
		w.WriteHeader(errors.STATUS_NOT_AUTHORIZED)
		return
	}

	if r.Method == http.MethodGet {
		var fileName string
		result = server.DownloadPicture(userId, userId, pictureId, &fileName)
		http.ServeContent(w, r, fileName, time.Now(), bytes.NewReader(result.Value))
		return
	} else if r.Method == http.MethodDelete {
		result = server.DeletePicture(userId, pictureId)

	} else {
		w.WriteHeader(errors.STATUS_METHOD_NOT_ALLOWED)
		return
	}

	w.WriteHeader(result.StatusCode)
	w.Write(result.Value)
}

// Dummy yet
func PictureFileHandler(w http.ResponseWriter, r *http.Request) {
	// result := &model.ReturnValue{}

	// Picture Download method
}
