package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/zeebo/gifstream"
)

type zeroCond struct {
	once sync.Once
	mu   *sync.Mutex
	cond *sync.Cond
}

func (z *zeroCond) init() {
	z.mu = new(sync.Mutex)
	z.cond = sync.NewCond(z.mu)
}

func (z *zeroCond) Lock() {
	z.once.Do(z.init)
	z.mu.Lock()
}

func (z *zeroCond) Unlock() {
	z.once.Do(z.init)
	z.mu.Unlock()
}

func (z *zeroCond) Wait() {
	z.once.Do(z.init)
	z.cond.Wait()
}

func (z *zeroCond) Broadcast() {
	z.once.Do(z.init)
	z.cond.Broadcast()
}

type RandomProvider struct {
	Delay   time.Duration
	Bounds  image.Rectangle
	Palette color.Palette

	once    sync.Once
	cond    zeroCond
	current *image.Paletted
}

func RandomColor() color.Color {
	return color.RGBA{
		R: byte(rand.Intn(256)),
		G: byte(rand.Intn(256)),
		B: byte(rand.Intn(256)),
		A: 255}
}

func (r *RandomProvider) Image() *image.Paletted {
	r.once.Do(r.start)

	r.cond.Lock()
	defer r.cond.Unlock()

	for r.current == nil {
		r.cond.Wait()
	}
	return r.current
}

func (r *RandomProvider) start() {
	r.cond.Lock()
	defer r.cond.Unlock()

	r.makeImage()
	r.cond.Broadcast()

	go func() {
		for {
			time.Sleep(r.Delay)
			r.makeImage()
		}
	}()
}

func (r *RandomProvider) makeImage() {
	p := image.NewPaletted(r.Bounds, r.Palette)
	min, max := r.Bounds.Min, r.Bounds.Max
	for x := min.X; x < max.X; x++ {
		col := RandomColor()
		for y := min.Y; y < max.Y; y++ {
			p.Set(x, y, col)
		}
	}
	r.current = p
}

func main() {
	s := gifstream.Streamer{
		Provider: &RandomProvider{
			Delay:  time.Second,
			Bounds: image.Rect(0, 0, 500, 500),
			Palette: color.Palette{
				RandomColor(), RandomColor(), RandomColor()}},
		Delay: time.Second}
	panic(http.ListenAndServe(":9999", http.HandlerFunc(func(
		w http.ResponseWriter, req *http.Request) {

		_, bw, err := w.(http.Hijacker).Hijack()
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Println(s.Stream(bw))
		}()
	})))
}
