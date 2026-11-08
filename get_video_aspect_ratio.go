package main

import (
	"bytes"
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
	random := "WORDS"
	return "", nil
}

type RatioResp struct {
	Streams []struct {
		Index  int `json:"index"`
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"streams"`
}
