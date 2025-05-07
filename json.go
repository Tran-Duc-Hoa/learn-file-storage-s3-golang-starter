package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func getMimeExtension(mediaType string) (string, error) {
	mimeExtensions := map[string]string{
		"image/png":        ".png",
		"image/jpeg":       ".jpg",
		"video/mp4":       ".mp4",
	}

	if ext, ok := mimeExtensions[mediaType]; ok {
		return ext, nil
	}
	return "", fmt.Errorf("unsupported media type: %s", mediaType)
}

func getVideoAspectRatio(filePath string) (string, error) {
	type Stream struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	type FFProbeOutput struct {
		Streams []Stream `json:"streams"`
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var result FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return "", fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	if len(result.Streams) == 0 {
		return "", fmt.Errorf("no streams found in video file")
	}

	width := result.Streams[0].Width
	height := result.Streams[0].Height

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid width or height in video stream")
	}

	ratio := float64(width) / float64(height)
	if ratio > 1.7 && ratio < 1.8 {
		return "16:9", nil
	} else if ratio > 0.55 && ratio < 0.57 {
		return "9:16", nil
	} else {
		return "other", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to process video for fast start: %w", err)
	}

	return outputFilePath, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	
	presignedRequest, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expireTime))

	if err != nil {
		return "", err
	}
	return presignedRequest.URL, nil
}




