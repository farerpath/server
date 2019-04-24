package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	api = "http://localhost/v1/"
)

var (
	token = ""
)

func main() {
	TestRegister()
	TestLogin()
	TestGetAlbumList()
	TestMakeAlbum()
	TestGetAlbumList()
}

type responseType struct {
	Checksum 	string
	Value 		map[string]interface{}
}

type P map[string]interface{}

func TestRegister() {
	payload, err := json.Marshal(P{"userID":"farerpath", "password":"Test1234$", "email":"farerpath@fp.com"})
	if err != nil {
		log.Fatalf("Register faild: %v", err)
	}

	resp, err := http.Post(api+"Register","application/json", bytes.NewReader(payload))
	if err != nil {
		log.Fatalf("Register faild: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Register failed: %v", resp.StatusCode)
	}
}

func TestLogin() {
	payload, err := json.Marshal(P{"userID":"farerpath", "password":"Test1234$"})
	if err != nil {
		log.Fatalf("Register faild: %v", err)
	}

	resp, err := http.Post(api+"Login", "application/json",bytes.NewReader(payload))
	if err != nil {
		log.Fatalf("Login faild: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Login failed: %v", resp.StatusCode)
	}

	respPayload := &responseType{}

	data, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(data, respPayload)

	token = respPayload.Value["authToken"].(string)
	if len(token) <1 {
		log.Fatalf("Token not valid")
	}

	fmt.Println(token)
}

func TestGetAlbumList() {
	req, err := http.NewRequest("GET", api + "Album/farerpath", nil)
	if err != nil {
		log.Fatalf("GetAlbumList faild: %v", err)
	}
	req.Header.Add("X-Farerpath-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("GetAlbumList faild: %v", err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("GetAlbumList failed: %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(data))
}

func TestMakeAlbum() {
	payload, err := json.Marshal(P{"albumName":"testAlbum123", "beginTime":strconv.FormatInt(time.Now().Unix(), 10), "members":""})
	if err != nil {
		log.Fatalf("Register faild: %v", err)
	}


	req, err := http.NewRequest("POST", api + "Album/farerpath", bytes.NewReader(payload))
	if err != nil {
		log.Fatalf("MakeAlbum faild: %v", err)
	}
	req.Header.Add("X-Farerpath-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("MakeAlbum faild: %v", err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("MakeAlbum failed: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(data))
}