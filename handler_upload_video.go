package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 10 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", errors.New("Not authorized"))
		return
	}
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, 500, "multipartform", err)
		return
	}
	mFile, mHeader, err := r.FormFile("video")

	if err != nil {
		respondWithError(w, 500, "couldn't make mFile", err)
		return
	}
	defer mFile.Close()

	mediaType, _, err := mime.ParseMediaType(mHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, 500, "Check media type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, 400, "Non matching content type", err)
		return
	}
	file, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, 500, "Couldn't create file", err)
		return
	}
	defer os.Remove(file.Name())
	defer file.Close()

	_, err = io.Copy(file, mFile)
	if err != nil {
		respondWithError(w, 500, "can't copy file", err)
		return
	}
	processedFilePath, err := processVideoForFastStart(file.Name())
	if err != nil {
		respondWithError(w, 500, "Couldn't process video", err)
		return
	}
	processedFile, err := os.Open(processedFilePath)
	if err != nil {
		respondWithError(w, 500, "Couldn't open processed file", err)
		return
	}
	defer os.Remove(processedFilePath)
	defer processedFile.Close()

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, 500, "can't seek", err)
		return
	}
	aspect, err := getVideoAspectRation(file.Name())
	if err != nil {
		respondWithError(w, 500, "Unable to get aspect ratio", err)
		return
	}
	switch aspect {
	case "16:9":
		aspect = "landscape"
	case "9:16":
		aspect = "portrait"
	default:
		aspect = "other"
	}
	encodedVideoString, err := getRandomURLEncoding()
	if err != nil {
		respondWithError(w, 500, "couldn't encode", err)
		return
	}
	key := fmt.Sprintf("%v/%v.mp4", aspect, encodedVideoString)
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &key,
		Body:        processedFile,
		ContentType: &mediaType,
	})
	if err != nil {
		respondWithError(w, 500, "Couldn't make put object", err)
		return
	}
	videoURLString := fmt.Sprintf("%v,%v", cfg.s3Bucket, key)
	video.VideoURL = &videoURLString
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, 500, "Couldn't update video", err)
		return
	}
	signedVideo, err := cfg.dbVideoToSignedVideo(video)
	if err != nil {
		respondWithError(w, 500, "Couldn't sign video URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, signedVideo)
}
