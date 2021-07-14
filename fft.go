package main

import (
	"fmt"
	"image/color"
	"math/cmplx"

	"gonum.org/v1/plot/plotter"

	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

func gonumFFT(wave []float64) {
	// Fast Fourier Transform
	fft := fourier.NewFFT(len(wave))
	coeff := fft.Coefficients(nil, wave)

	pts := make(plotter.XYs, len(coeff))
	for i, c := range coeff {
		pts[i].X = fft.Freq(i)
		pts[i].Y = cmplx.Abs(c)
		// fmt.Printf("freq=%v cycles/period, magnitude=%v, phase=%.4g\n",
		// 	fft.Freq(i), cmplx.Abs(c), cmplx.Phase(c))
	}

	// Draw Plot
	// https://godoc.org/gonum.org/v1/plot/plotter#Line
	p, err := plot.New()
	if err != nil {
		fmt.Printf("plotter Error : %v", err)
	}

	p.Title.Text = "Fourier transform depicting the frequency components"
	p.X.Label.Text = "Frequency"
	p.Y.Label.Text = "Amplitude"
	p.Add(plotter.NewGrid())

	filled, err := plotter.NewLine(pts)
	if err != nil {
		fmt.Printf("plotter Error : %v", err)
	}
	filled.Color = color.RGBA{R: 76, G: 225, B: 223, A: 0}
	p.Add(filled)

	// Save the plot to a PNG file.
	if err := p.Save(8*vg.Inch, 4*vg.Inch, "data/output.png"); err != nil {
		fmt.Printf("plotter Error : %v", err)
	}
}
