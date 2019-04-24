package server

import (
	"log"
	"net/http"
	"regexp"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	albumService "github.com/farerpath/albumservice/proto"
	authService "github.com/farerpath/authservice/proto"

	"github.com/farerpath/server/model/errors"
	"github.com/farerpath/server/model/model"

	"github.com/farerpath/server/common/util"
)

func Login(loginUserId string, loginPassword string, loginDeviceType int, loginIP string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	loginRequest := authService.LoginRequest{
		LoginUserID:     loginUserId,
		LoginPassword:   loginPassword,
		LoginDeviceType: int32(loginDeviceType),
		LoginIP:         loginIP,
	}

	loginResp, err := authClient.Login(context.Background(), &loginRequest)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
		} else if st.Code() == codes.Unauthenticated {
			returnValue.StatusCode = http.StatusUnauthorized
		} else {
			log.Printf("Unable to login. GRPC Service returned error.\n%v", err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		}

		return
	}

	loginReturnValue := &model.LoginReturnValue{
		UserName:     loginResp.GetUserName(),
		UserID:       loginResp.GetUserID(),
		Country:      loginResp.GetCountry(),
		RefreshToken: loginResp.GetRefreshToken(),
		AuthToken:    loginResp.GetAuthToken(),
	}

	returnValue.Value = util.MakeReturnValueToJson(loginReturnValue)
	return
}

func TokenLogin(token, userID string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	resp, err := authClient.GetUser(context.Background(), &authService.GetUserRequest{UserID: userID})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
		}
		log.Println(err)
		returnValue.StatusCode = http.StatusInternalServerError
		return
	}

	returnValue.Value = util.MakeReturnValueToJson(map[string]interface{}{"user_id": resp.GetUserID(), "user_name": resp.GetUserName(), "country": resp.GetCountry(), "token": token})
	return
}

func Logout(token string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	_, err := authClient.Logout(context.Background(), &authService.LogoutRequest{Token: token})
	if err != nil {
		log.Printf("Unable to login. GRPC Service returned error.\n%v", err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	return
}

func Register(userId, email, password string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	if !(isEmailValid(email) && isUserIDValid(userId) && isPasswordValid(password)) {
		returnValue.StatusCode = http.StatusBadRequest
		return
	}

	req := &authService.RegisterRequest{
		Email:    email,
		UserID:   userId,
		Password: password,
	}

	_, err := authClient.Register(context.Background(), req)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.AlreadyExists {
			returnValue.StatusCode = http.StatusConflict
		} else {
			log.Printf("Unable to Register. GRPC Service returned error.\n%v", err)
			returnValue.StatusCode = http.StatusInternalServerError
		}

		return
	}

	_, err = albumClient.MakeAlbumList(context.Background(), &albumService.MakeAlbumListRequest{ReqUserID: userId})
	if err != nil {
		log.Printf("Unable to Register. Failed to make empty album list GRPC Service returned error.\n%v", err)
		returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
		return
	}

	return
}

func isEmailValid(email string) bool {
	// Check Email address
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !re.MatchString(email) {
		return false
	}

	return true
}

func isUserIDValid(userID string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9]*$")
	if !re.MatchString(userID) {
		return false
	}

	return true
}

func isPasswordValid(password string) bool {
	if len(password) < 6 {
		return false
	}

	return true
}
