package server

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/farerpath/server/common/util"
	"github.com/farerpath/server/model/errors"
	"github.com/farerpath/server/model/model"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authService "github.com/farerpath/authservice/proto"
)

func GetUserInformation(sessionUserId, userId string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	resp, err := authClient.GetUser(context.Background(), &authService.GetUserRequest{UserID: userId})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			returnValue.StatusCode = http.StatusNotFound
			return
		}

		log.Println(err)
		returnValue.StatusCode = http.StatusInternalServerError
		return

		return
	}

	result := &model.Account{
		Email:            resp.GetEmail(),
		UserID:           resp.GetUserID(),
		UserName:         resp.GetUserName(),
		FirstName:        resp.GetFirstName(),
		LastName:         resp.GetLastName(),
		Country:          resp.GetCountry(),
		ProfilePhotoPath: resp.GetProfilePhotoPath(),
		SessionDuration:  resp.GetSessionDuration(),
		IsBirthdayPublic: resp.GetIsBirthdayPublic(),
		IsCountryPublic:  resp.GetIsCountryPublic(),
		IsProfilePublic:  resp.GetIsProfilePublic(),
	}

	if sessionUserId != userId {
		if !resp.GetIsProfilePublic() {
			returnValue.StatusCode = http.StatusForbidden
			return
		}

		result.SessionDuration = 0

		if !resp.GetIsBirthdayPublic() {
			result.Birthday = time.Time{}
		}

		if !resp.GetIsCountryPublic() {
			result.Country = ""
		}

		returnValue.Value = util.MakeReturnValueToJson(result)
	}

	returnValue.Value = util.MakeReturnValueToJson(result)

	return
}

func UpdateUserPassword(sessionUserId, userId string, oldPassword, newPassword string) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	if sessionUserId != userId {
		returnValue.StatusCode = errors.STATUS_NOT_AUTHORIZED
		return
	}

	if !isPasswordValid(newPassword) {
		returnValue.StatusCode = http.StatusBadRequest
		return
	}

	/*
		resp, err := dbClient.FindUser(context.Background(), &dbService.UserId{UserId: userId})
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
			return
		}

		if !util.IsSucced(resp.StatusCode) {
			returnValue.StatusCode = int(resp.GetStatusCode())
			return
		}

		user := &dbService.User{}
		err = ptypes.UnmarshalAny(resp.Value, user)
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_FARERPATH_ERROR
			returnValue.Value = []byte(string(errors.FP_PASSWORD_NOT_MATCH))
			return
		}

		newPsw, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
			return
		}

		user.Password = string(newPsw)

		resp, err = dbClient.UpdateUser(context.Background(), &dbService.NewUser{UserId: userId, NewUser: user})
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
			return
		}

		if !util.IsSucced(resp.StatusCode) {
			returnValue.StatusCode = int(resp.GetStatusCode())
			return
		}
	*/
	return
}

func UpdateUserProfile(sessionUserID, userID, userName, email, profilePhotoPath, country, birthday string, isCountryPublic, isBirthdayPublic, isProfilePublic bool) (returnValue *model.ReturnValue) {
	returnValue = util.InitReturnValue()

	if sessionUserID != userID {
		returnValue.StatusCode = errors.STATUS_NOT_AUTHORIZED
		return
	}

	/*
		resp, err := dbClient.FindUser(context.Background(), &dbService.UserId{UserId: userID})
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
			return
		}

		if !util.IsSucced(resp.StatusCode) {
			returnValue.StatusCode = int(resp.GetStatusCode())
			return
		}

		user := &dbService.User{}
		err = ptypes.UnmarshalAny(resp.Value, user)
		if err != nil {
			fmt.Println(err)
			returnValue.StatusCode = errors.STATUS_INTERNAL_ERROR
			return
		}

		if isUserNameValid(userName) {
			user.UserName = userName
		}

		if isEmailValid(email) {
			user.Email = email
		}

		if isCountryCodeValid(country) {
			user.Country = country
		}

		if isBirthdayValid(birthday) {
			user.Birthday = birthday
		}

		//user.ProfilePhotoPath = profilePhotoPath
		user.IsCountryPublic = isCountryPublic
		user.IsBirthdayPublic = isBirthdayPublic
		user.IsProfilePublic = isProfilePublic
	*/

	return
}

func isUserNameValid(userName string) bool {
	if len(userName) < 3 {
		return false
	}

	re := regexp.MustCompile("^[a-zA-Z0-9]*$")
	return re.MatchString(userName)
}

func isCountryCodeValid(code string) bool {
	if len(code) != 2 || len(code) != 3 {
		return false
	}

	return true
}

func isBirthdayValid(birthday string) bool {
	if len(birthday) != 8 {
		return false
	}

	return true
}
