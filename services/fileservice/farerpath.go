package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/farerpath/server/model/consts"
)

func App() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", DownloadHandler).Methods(http.MethodGet)
	r.HandleFunc("/", UploadHandler).Methods(http.MethodPost)
	r.HandleFunc("/", DeleteHandler).Methods(http.MethodDelete)

	n := negroni.New(negroni.NewLogger(), negroni.NewRecovery())
	n.UseHandler(r)

	return n
}

func main() {
	port := flag.String("port", "80", "bind port (default: 80)")

	err := http.ListenAndServe(":"+*port, App())
	log.Fatal(err)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.FormValue("userId")
	pictureId := r.FormValue("pictureId")

	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	awsCredential := credentials.NewStaticCredentials(consts.AWS_ACCESS_KEY, consts.AWS_SECRET_ACCESS_KEY, "")
	awsSession := session.Must(session.NewSession(&aws.Config{Credentials: awsCredential, Region: aws.String(endpoints.ApNortheast2RegionID)}))

	awsS3 := s3.New(awsSession)

	key := userId + "/" + pictureId

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = awsS3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("farerpathalpha1"),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buffer),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	buffer := &aws.WriteAtBuffer{}

	awsCredential := credentials.NewStaticCredentials(consts.AWS_ACCESS_KEY, consts.AWS_SECRET_ACCESS_KEY, "")
	awsSession := session.Must(session.NewSession(&aws.Config{Credentials: awsCredential, Region: aws.String(endpoints.ApNortheast2RegionID)}))

	key := r.FormValue("userId") + "/" + r.FormValue("pictureId")

	downloader := s3manager.NewDownloader(awsSession)

	downloader.Download(buffer, &s3.GetObjectInput{
		Bucket: aws.String("farerpathalpha1"),
		Key:    aws.String(key),
	})

	http.ServeContent(w, r, r.FormValue("pictureId"), time.Now(), bytes.NewReader(buffer.Bytes()))
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("userId") + "/" + r.FormValue("pictureId")

	ns, err := session.NewSession()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	svc := s3.New(ns)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String("farerpathalpha1"),
		Key:    aws.String(key),
	}

	_, err = svc.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
