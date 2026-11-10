package main

import "os/exec"

func processVideoForFastStart(filepath string) (string, error) {
	processingFilePath := filepath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", processingFilePath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return processingFilePath, nil
}
