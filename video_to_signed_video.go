package main

import (
	"errors"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	bucketAndKey := strings.Split(*video.VideoURL, ",")
	if len(bucketAndKey) != 2 {
		return database.Video{}, errors.New("The bucket and key weren't retrievable")
	}
	bucket := bucketAndKey[0]
	key := bucketAndKey[1]
	signedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute*5)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &signedURL

	return video, nil
}
