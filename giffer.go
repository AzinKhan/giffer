package giffer

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"log"
	"sync"
)

func decodeJPG(data []byte) (image.Image, error) {
	log.Println("Decoding JPEG")
	img, err := jpeg.Decode(bytes.NewReader(data))
	return img, err
}

func ConvertToGIF(img image.Image) (*image.Paletted, error) {
	log.Println("Encoding to GIF")
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

func Convert(data []byte) (*image.Paletted, error) {
	img, err := decodeJPG(data)
	if err != nil {
		log.Printf("Error decoding JPEG: %+v", err)
		return nil, err
	}
	GIF, err := ConvertToGIF(img)
	if err != nil {
		log.Printf("Error converting to GIF: %+v", err)
		return nil, err
	}
	return GIF, nil

}

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
	log.Printf("Encoding %+v images into GIF", len(G.Image))
	var buf []byte
	Buf := bytes.NewBuffer(buf)
	err := gif.EncodeAll(Buf, G)
	return Buf, err
}
