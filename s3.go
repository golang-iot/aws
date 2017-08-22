package aws

import (
	"os"
	"log"
	"time"
	"bytes"
	"net/http"
	"path/filepath"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil" 
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

type s3Uploader interface{
	init()
	Put(file string, path string, bucket string) error
	SendToRekognition(file string) ([]Face, error)
}

type S3Manager struct{
	S3Service *s3.S3
	RekognitionService *rekognition.Rekognition
}

func (s *S3Manager) Init(aws_access_key_id string, aws_secret_access_key string, token string, region string){
	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, token)
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	s.S3Service = s3.New(session.New(), cfg)
	log.Printf("Connecting to AWS on %v",cfg)
	s.RekognitionService = rekognition.New(session.New(), cfg)
}

func (s S3Manager) Put(filename string, path string, bucket string) error{
	file, err := os.Open(filename) 
	if err != nil { 
	  log.Printf("err opening file: %s", err) 
	} 
	defer file.Close()
	
	fileInfo, _ := file.Stat() 
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size) 
	file.Read(buffer) 
	fileBytes := bytes.NewReader(buffer) 
	fileType := http.DetectContentType(buffer)
	
	bucketPath := path+"/"+filepath.Base(file.Name())

	params := &s3.PutObjectInput{ 
	  Bucket: aws.String(bucket), 
	  Key: aws.String(bucketPath), 
	  Body: fileBytes, 
	  ContentLength: aws.Int64(size), 
	  ContentType: aws.String(fileType), 
	}
	resp, err := s.S3Service.PutObject(params) 
	if err != nil { 
	  log.Printf("bad response: %s", err) 
	} 
	log.Printf("response %s", awsutil.StringValue(resp))
	return err
}

func (s S3Manager) SendToRekognition(filename string) ([]Face, error){
	file, err := os.Open(filename) 
	if err != nil { 
	  log.Printf("err opening file: %s", err) 
	} 
	defer file.Close()
	
	fileInfo, _ := file.Stat() 
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size) 
	file.Read(buffer) 
		
	img := rekognition.Image{Bytes:buffer}
	
	attrs:= []*string{aws.String("ALL")}
	
	var faces []Face
	
	input := rekognition.DetectFacesInput{Image:&img,Attributes:attrs}
	output, err := s.RekognitionService.DetectFaces(&input)
	for i, fd := range output.FaceDetails {
        log.Printf("The person %v is ", i)
		log.Printf("-%v",fd.AgeRange)
		log.Printf("-%v",fd.Gender)
		log.Printf("-Smile? %v",fd.Smile)
		log.Printf("Emotions:")
		
		var emotions []string
		
        for _, e := range fd.Emotions {
                log.Printf("%v, ", *e.Type)
				emotions = append(emotions,*e.Type)
        }
        
		now := time.Now()
		f := Face{
			MaxAge: *fd.AgeRange.High,
			MinAge: *fd.AgeRange.Low,
			Gender: *fd.Gender.Value,
			GenderConf: *fd.Gender.Confidence,
			Smile: *fd.Smile.Value,
			SmileConf: *fd.Smile.Confidence,
			Emotions: emotions,
			Created: now}
		
		faces = append(faces,f)
	}
	
	return faces, err
	
}