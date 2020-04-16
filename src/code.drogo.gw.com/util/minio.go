package util

import (
	"github.com/minio/minio-go"
	"log"
	"strings"
	"regexp"
	"encoding/base64"
	"github.com/satori/go.uuid"
	"errors"
)

type MinioClient struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
}

func NewMinioClient() MinioClient {
	return MinioClient{endpoint: "minio.eventackle.com", accessKeyID: "TDMWDI8Z62BHU582XRPZ", secretAccessKey: "OQmshdAimF5pLLrRyLjA6/Pcr6WidSSNXJjhOtOE"}
}

func (mc MinioClient) StoreImage(name string, imageData string) (string, error) {

	minioClient, err := minio.New(mc.endpoint, mc.accessKeyID, mc.secretAccessKey, false)
	if err != nil {
		log.Fatalln(err)
	}
	bucketName := "uploads"
	re := regexp.MustCompile("[:;,]")
	result := re.Split(imageData, -1)
	fileType := strings.Split(result[1], "/")
	_uuid, _ := uuid.NewV4()
	nameWithoutSpaces := strings.Replace(name, " ", "_", -1)
	objectName := _uuid.String() + "_" + nameWithoutSpaces + "." + fileType[1]
	decodedImage, _ := base64.StdEncoding.DecodeString(result[3])
	imgData := strings.NewReader(string(decodedImage))

	_, err = minioClient.PutObject(bucketName, objectName, imgData, -1, minio.PutObjectOptions{ContentType: result[1]})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func (mc MinioClient) DeleteImage(uploadId string) error {
	minioClient, err := minio.New(mc.endpoint, mc.accessKeyID, mc.secretAccessKey, false)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	bucketName := "uploads"
	return minioClient.RemoveObject(bucketName, uploadId)
}

func StorePDF(agreementData string) (string, error) {
	if agreementData == "" {
		return "", errors.New("empty agreement upload_id")
	}
	S3_ENDPOINT := "minio.eventackle.com"
	S3_KEY := "TDMWDI8Z62BHU582XRPZ"
	S3_SECRET := "OQmshdAimF5pLLrRyLjA6/Pcr6WidSSNXJjhOtOE"


	minioClient, err := minio.New(S3_ENDPOINT, S3_KEY, S3_SECRET, false)
	if err != nil {
		return "", err
	}

	bucketName := "uploads"
	re := regexp.MustCompile("[:;,]")
	result := re.Split(agreementData, -1)
	fileType := strings.Split(result[1], "/")
	_uuid, _ := uuid.NewV4()
	objectName := _uuid.String() + "Agreement." + fileType[1]
	decodedImage, _ := base64.StdEncoding.DecodeString(result[3])
	imgData := strings.NewReader(string(decodedImage))

	_, err = minioClient.PutObject(bucketName, objectName, imgData, -1, minio.PutObjectOptions{ContentType: result[1]})
	if err != nil {
		return "", err
	}
	return objectName, nil
}
