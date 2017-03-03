package main

import (
	"github.com/disintegration/imaging"
	"log"
	"time"
	"sync"
	"github.com/disintegration/gift"
	"image/color"
	"image"
	"os"
)

func main() {
	src, err := imaging.Open(os.Getenv("GOPATH") + "/src/github.com/supme/goElFinder/example/speedTest/img/Горы.jpg")
	if err != nil {
		log.Print(err)
	}

	n := 10
	w := 640
	d := float32(45)
	log.Printf("Repeat count %d. Resize width %d. Rotate degree %d", n, w, d)

	start := time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.NearestNeighbor)
	}
	elapsed := time.Since(start)
	log.Printf("Resize NearestNeighbor %s", elapsed)

	start = time.Now()
	s := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.NearestNeighbor)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize NearestNeighbor %s", elapsed)

	start = time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.Box)
	}
	elapsed = time.Since(start)
	log.Printf("Resize Box %s", elapsed)

	start = time.Now()
	s = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.Box)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize Box %s", elapsed)

	start = time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.Linear)
	}
	elapsed = time.Since(start)
	log.Printf("Resize Linear %s", elapsed)

	start = time.Now()
	s = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.Linear)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize Linear %s", elapsed)

	start = time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.MitchellNetravali)
	}
	elapsed = time.Since(start)
	log.Printf("Resize MitchellNetravali %s", elapsed)

	start = time.Now()
	s = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.MitchellNetravali)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize MitchellNetravali %s", elapsed)

	start = time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.CatmullRom)
	}
	elapsed = time.Since(start)
	log.Printf("Resize CatmullRom %s", elapsed)

	start = time.Now()
	s = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.CatmullRom)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize CatmullRom %s", elapsed)

	start = time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.Gaussian)
	}
	elapsed = time.Since(start)
	log.Printf("Resize Gaussian %s", elapsed)

	start = time.Now()
	s = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.Gaussian)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize Gaussian %s", elapsed)

	start = time.Now()
	for i := 0; i < n; i++ {
		imaging.Resize(src, w, 0, imaging.Lanczos)
	}
	elapsed = time.Since(start)
	log.Printf("Resize Lanczos %s", elapsed)

	start = time.Now()
	s = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		s.Add(1)
		go func() {
			imaging.Resize(src, w, 0, imaging.Lanczos)
			s.Done()
		}()
	}
	s.Wait()
	elapsed = time.Since(start)
	log.Printf("Parallel resize Lanczos %s", elapsed)


	gf := gift.New(gift.Rotate(d, color.Black, gift.NearestNeighborInterpolation))
	start = time.Now()
	for i := 0; i < n; i++ {
		dst := image.NewRGBA(gf.Bounds(src.Bounds()))
		gf.Draw(dst, src)
	}
	elapsed = time.Since(start)
	log.Printf("Rotate Nearest Neighbor Interpolation %s", elapsed)

	gf = gift.New(gift.Rotate(d, color.Black, gift.LinearInterpolation))
	start = time.Now()
	for i := 0; i < n; i++ {
		dst := image.NewRGBA(gf.Bounds(src.Bounds()))
		gf.Draw(dst, src)
	}
	elapsed = time.Since(start)
	log.Printf("Rotate Linear Interpolation %s", elapsed)

	gf = gift.New(gift.Rotate(d, color.Black, gift.CubicInterpolation))
	start = time.Now()
	for i := 0; i < n; i++ {
		dst := image.NewRGBA(gf.Bounds(src.Bounds()))
		gf.Draw(dst, src)
	}
	elapsed = time.Since(start)
	log.Printf("Rotate Cubic Interpolation %s", elapsed)

}