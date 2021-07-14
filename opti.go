package main

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"gocv.io/x/gocv"
)

type Params struct {
	maxCorners   int
	qualityLevel float64
	minDistance  float64
	frameIdx     int
	detectFrame  int
}

func setVecbAt(m gocv.Mat, row int, col int, v gocv.Vecf) {
	ch := m.Channels()
	// https://github.com/hybridgroup/gocv/issues/339
	for c := 0; c < ch; c++ {
		m.SetFloatAt(row, col*ch+c, v[c])
	}
}

func calOptcialFlow(videoPath string, ch chan gocv.Mat) {

	cap, err := gocv.VideoCaptureFile(videoPath)
	defer cap.Close()
	if err != nil {
		fmt.Printf("Invalid Video file : %v\n", videoPath)
		return
	}

	p := Params{maxCorners: 500, qualityLevel: 0.01, minDistance: 5, frameIdx: 0, detectFrame: 10000000}
	// kernel := gocv.GetStructuringElement(0, image.Pt(10, 10))

	rio := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: int(cap.Get(3)), Y: int(cap.Get(4))},
	}

	curr := gocv.NewMat()
	defer curr.Close()
	currGray := gocv.NewMat()
	defer currGray.Close()
	prevGray := gocv.NewMat()
	defer prevGray.Close()

	p0 := gocv.NewMat()
	defer p0.Close()
	p1 := gocv.NewMat()
	defer p1.Close()
	st := gocv.NewMat()
	defer st.Close()
	e := gocv.NewMat()
	defer e.Close()

	var point []float64

	for {
		if ok := cap.Read(&curr); !ok {
			break
		}

		curr = curr.Region(rio)
		gocv.CvtColor(curr, &currGray, gocv.ColorBGRToGray)
		gocv.Threshold(currGray, &currGray, 127, 255, gocv.ThresholdBinary)
		// gocv.Erode(currGray, &currGray, kernel)

		if !p0.Empty() {

			// Calculate Optical Flow
			gocv.CalcOpticalFlowPyrLK(prevGray, currGray, p0, p1, &st, &e)
			if st.Empty() { // st shape [50 1 1]
				fmt.Printf("Error in CalcOpticalFlowPyrLK test")
				continue
			}
			if st.Cols() != 1 {
				fmt.Printf("Invalid CalcOpticalFlowPyrLK test cols: %v", st.Cols())
				return
			}
			// Select good points
			goodNew := gocv.NewMatWithSize(p1.Rows(), p1.Cols(), p1.Type())
			goodOld := gocv.NewMatWithSize(p0.Rows(), p0.Cols(), p0.Type())

			for r := 0; r < p1.Rows(); r++ { // p1 shape [100 1 2]
				if st.GetUCharAt(r, 0) == 1 {
					// goodNew.SetUCharAt(i, j*nextPts.Channels()+1, nextPts.GetUCharAt(i, j))
					vec := p1.GetVecfAt(r, 0)
					setVecbAt(goodNew, r, 0, vec)
				}
			}

			for r := 0; r < p0.Rows(); r++ {
				if st.GetUCharAt(r, 0) == 1 {
					vec := p0.GetVecfAt(r, 0)
					setVecbAt(goodOld, r, 0, vec)
				}
			}
			// Draw and Show
			for i := 0; i < goodNew.Rows(); i++ {
				v0 := goodOld.GetVecfAt(i, 0)
				v1 := goodNew.GetVecfAt(i, 0)
				a := v1[0] - v0[0] // currX - prevX
				// b := v0[1] - v1[1] // y
				// d := math.Sqrt(float64((a * a) + (b * b)))

				gocv.Circle(&curr, image.Pt(int(v1[0]), int(v1[1])), 1, color.RGBA{190, 233, 218, 0}, 5)
				gocv.Line(&curr, image.Pt(int(v0[0]), int(v0[1])), image.Pt(int(v1[0]), int(v1[1])), color.RGBA{226, 115, 196, 0}, 2)
				gocv.PutText(&curr, strconv.Itoa(i), image.Pt(int(v1[0]), int(v1[1])), gocv.FontHersheySimplex, 0.5, color.RGBA{250, 70, 77, 0}, 1)
				if i == 299 { //num of point
					point = append(point, float64(a))
				}
			}
			goodNew.CopyTo(&p0) //corners = nextPts

			goodOld.Close()
			goodNew.Close()
		}
		// Send gocv.Mat
		ch <- curr
		// ch <- currGray
		// Update
		if (p.frameIdx % p.detectFrame) == 0 {
			gocv.GoodFeaturesToTrack(currGray, &p0, p.maxCorners, p.qualityLevel, p.minDistance)
			tc := gocv.NewTermCriteria(gocv.Count|gocv.EPS, 30, 0.03)
			gocv.CornerSubPix(currGray, &p0, image.Pt(10, 10), image.Pt(-1, -1), tc)
			// 서브 픽셀 코너 검출기는 많은 (전형적으로 20 내지 100) 포인트가 검출되는 이미지에서 사용되도록 설계)
		}
		p.frameIdx++
		currGray.CopyTo(&prevGray)
	}
	// Send point data to FFT
	gonumFFT(point)
}

// 코너 검출 OpenCV
// https://deep-learning-study.tistory.com/251
