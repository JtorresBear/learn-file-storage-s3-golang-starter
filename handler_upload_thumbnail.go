package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse", err)
		return
	}

	mFile, mHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get FormFile", err)
		return
	}
	defer mFile.Close()

	contentType := mHeader.Header.Get("Content-Type")

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No video found", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusForbidden, "forbidden", errors.New("forbidden"))
		return
	}

	extSlice, err := mime.ExtensionsByType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	if len(extSlice) < 1 {
		respondWithError(w, http.StatusBadRequest, "no extension found", errors.New("no extension found"))
		return
	}

	//redo video path
	encodedVideoString, err := getRandomURLEncoding()
	if err != nil {
		respondWithError(w, 500, "couldn't encode", err)
		return
	}
	videoPath := filepath.Join(cfg.assetsRoot, encodedVideoString+extSlice[0])

	new_F, err := os.Create(videoPath)
	if err != nil {
		respondWithError(w, 500, "Can't create file", err)
		return
	}
	defer new_F.Close()

	_, err = io.Copy(new_F, mFile)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "problem writing file", err)
		return
	}
	err = new_F.Close()
	if err != nil {
		respondWithError(w, 500, "couldn't close file", err)
		return
	}
	thumbnail_url_s := fmt.Sprintf("http://localhost:%v/%v", cfg.port, videoPath)
	video.ThumbnailURL = &thumbnail_url_s
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, 500, "couldn't update video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, video)
}

func getRandomURLEncoding() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	encodedURL := base64.RawURLEncoding.EncodeToString(key)

	return encodedURL, nil
}
