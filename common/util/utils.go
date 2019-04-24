package util

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/farerpath/server/model/consts"
	"github.com/farerpath/server/model/model"

	"github.com/go-redis/redis"
)

func Logger(contents ...string) {
	var content string

	for _, s := range contents {
		content += s
	}

	fmt.Println(time.Now(), content)
}

func UnmarshalBody(body *io.ReadCloser) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	tmp, err := ioutil.ReadAll(*body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tmp, &m)
	return m, err
}

func MakeReturnValueToJson(returnValue model.Any) []byte {
	var value model.ResponseValue

	value.Value = returnValue
	value.Checksum = HashAndEncodeString(consts.APPSECRET)

	tmp, _ := json.Marshal(value)
	return tmp
}

func IsStringVulnerable(strs ...string) bool {
	str := ""

	for _, s := range strs {
		str += s
	}

	return strings.ContainsAny(str, "$ & { & } & [ & ] & ( & )")
}

func HashAndEncodeString(str ...string) string {
	h := sha512.New()
	var tmpStr string

	for _, s := range str {
		tmpStr += s
	}

	h.Write([]byte(tmpStr))
	bs := h.Sum([]byte{})

	return base64.StdEncoding.EncodeToString(bs)
}

func HashAndEncodeByte(bytes []byte) string {
	h := sha512.New()

	h.Write(bytes)
	bs := h.Sum([]byte{})

	return base64.StdEncoding.EncodeToString(bs)
}

func EncodeLoginSessionToJson(session *model.LoginSession) string {
	tmp, _ := json.Marshal(session)
	return string(tmp[:])
}

func IsLoggedIn(token string) (bool, string) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	val, err := client.Get(token).Result()
	if err == redis.Nil { // 로그인 안됨
		return false, ""
	} else if err != nil { // 오류 발생
		fmt.Println(err.Error())
		return false, ""
	} else {
		return true, val
	}
}

func DelSpaces(str string) string {
	return strings.Replace(str, " ", "", -1)
}

func InitReturnValue() (returnValue *model.ReturnValue) {
	returnValue = new(model.ReturnValue)

	returnValue.Value = nil
	returnValue.StatusCode = 200
	return
}

func FindInSlice(target string, slice []string) bool {
	for _, str := range slice {
		if target == str {
			return true
		}
	}

	return false
}

func IsSucced(statusCode int32) bool {
	return statusCode/100 == 2
}

func GetReqIP(ip string) string {
	return strings.Split(ip, ",")[0]
}
