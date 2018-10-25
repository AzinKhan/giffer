/*
Package giffer implements a small library for taking a series of images
and encoding them into a single GIF animation.
*/
package giffer

import (
	"bytes"
	"image"
	"image/gif"
	_ "image/jpeg" // Register JPEG decoder with image package
	_ "image/png"
	"log"
	"sync"
)

func decode(data []byte) (image.Image, string, error) {
	return image.Decode(bytes.NewReader(data))
}

// ConvertToGIF takes an image.Image and converts it into a GIF,
// returning an *image.Paletted.
func ConvertToGIF(img image.Image) (*image.Paletted, error) {
	var b []byte
	bf := bytes.NewBuffer(b)
	var opt gif.Options
	opt.NumColors = 256
	err := gif.Encode(bf, img, &opt)
	// Only way this returns an error seems to be if the image is too large
	if err != nil {
		return nil, err
	}
	im, err := gif.Decode(bf)
	return im.(*image.Paletted), err
}

var wg sync.WaitGroup

// Convert is a wrapper for ConvertToGIF, taking in a slice of bytes
// and returning a GIF encoded *image.Paletted.
func Convert(data []byte) (*image.Paletted, error) {
	img, kind, err := decode(data)
	if err != nil {
		log.Printf("Error decoding: %+v", err)
		return nil, err
	}
	log.Printf("Converting file type %s to GIF", kind)
	return ConvertToGIF(img)

}

// Giffer contains the main logic for this package, taking a series of
// byte slices, which are assumed to be images, and converting them into
// one GIF animation.
func Giffer(inputData [][]byte) (*bytes.Buffer, error) {
	G := &gif.GIF{
		LoopCount: 0,
		Disposal:  nil,
		Delay:     make([]int, len(inputData)),
		Image:     make([]*image.Paletted, len(inputData)),
	}
	log.Println("Converting images to GIF")
	errChan := make(chan error, len(inputData))
	wg.Add(len(inputData))
	for index, d := range inputData {
		data := d
		i := index
		go func() {
			defer wg.Done()
			GIF, err := Convert(data)
			errChan <- err
			G.Delay[i] = 8
			G.Image[i] = GIF
		}()
	}
	wg.Wait()
	// Don't forget to close the channel!
	close(errChan)
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}
	log.Printf("Combining %+v images into GIF", len(G.Image))
	var buf []byte
	Buf := bytes.NewBuffer(buf)
	err := gif.EncodeAll(Buf, G)
	return Buf, err
}
