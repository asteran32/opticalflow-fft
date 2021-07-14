package main

import (
	"time"

	"gocv.io/x/gocv"
)

func main() {

	videoPath := "data/ipcam03-2021-07-12_13-30.mp4"
	win := gocv.NewWindow("OpticalFlow Lucas Kaneda")
	defer win.Close()

	ch := make(chan gocv.Mat)

	go calOptcialFlow(videoPath, ch)

readFrame:
	for {
		select {
		case img := <-ch:
			win.IMShow(img)
			win.WaitKey(1) // ms

		case <-time.After(3 * time.Second):
			break readFrame
		}
	}
}
