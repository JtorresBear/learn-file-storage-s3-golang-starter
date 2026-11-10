package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

func getVideoAspectRation(filepath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	//nil is a valid 0 form of a byte slice=[]byte, so i used it to initialize a bytes.Buffer pointer
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	var results RatioResp

	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		return "", err
	}
	if len(results.Streams) < 1 {
		return "", err
	}
	if results.Streams[0].Height < 1 {
		return "", err
	}
	ratio := float64(results.Streams[0].Width) / float64(results.Streams[0].Height)

	switch {
	case ratio > 1.6 && ratio < 1.8:
		return "16:9", nil
	case ratio > 0.55 && ratio < 0.57:
		return "9:16", nil
	default:
		return "other", nil
	}
}

type RatioResp struct {
	Streams []struct {
		Index  int `json:"index"`
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"streams"`
}
