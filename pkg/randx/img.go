package randx

import (
	"bytes"
	"github.com/pbnjay/pixfont"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// GenerateRandomImageFile generate image file.
// If fastMode is true, a sparse file is filled with zero (ascii NUL) and doesn't actually take up the disk space
// until it is written to, but reads correctly.
// $ ls -lh 424661641.png
// -rw-------  1 bingoobjca  staff   488K Mar 15 12:19 424661641.png
// $ du -hs 424661641.png
// 8.0K    424661641.png
// If fastMode is false, an actually sized file will generated.
// $ ls -lh 1563611881.png
// -rw-------  1 bingoobjca  staff   488K Mar 15 12:28 1563611881.png
// $ du -hs 1563611881.png
// 492K    1563611881.png

type ImgConfig struct {
	Width      int
	Height     int
	RandomText string
	FastMode   bool
	PixelSize  int
}

func (c *ImgConfig) GenFile(filename string, fileSize int) int {
	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	if c.PixelSize == 0 {
		c.PixelSize = 40
	}

	data, imgSize := c.Gen(filepath.Ext(filename))
	f.Write(data)
	if fileSize <= imgSize {
		return imgSize
	}

	if !c.FastMode {
		b := Bytes(fileSize - imgSize)
		f.Write(b)
		return fileSize
	}

	// refer to https://stackoverflow.com/questions/16797380/how-to-create-a-10mb-file-filled-with-000000-data-in-golang
	// use f.Truncate to change size of the file
	// If you are using unix, then you can create a sparse file very quickly.
	// A sparse file is filled with zero (ascii NUL) and doesn't actually take up the disk space
	// until it is written to, but reads correctly.
	f.Truncate(int64(fileSize))
	return fileSize
}

// Gen generate a random image with imageFormat (jpg/png) .
// refer: https://onlinejpgtools.com/generate-random-jpg
func (c *ImgConfig) Gen(imageFormat string) ([]byte, int) {
	var img draw.Image

	format := strings.ToLower(imageFormat)
	switch format {
	case ".jpg", ".jpeg":
		img = image.NewNRGBA(image.Rect(0, 0, c.Width, c.Height))
	default: // png
		img = image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	}

	yp := c.Height / c.PixelSize
	xp := c.Width / c.PixelSize
	for yi := 0; yi < yp; yi++ {
		for xi := 0; xi < xp; xi++ {
			drawPixelWithColor(img, yi, xi, c.PixelSize, Color())
		}
	}

	if c.RandomText != "" {
		pixfont.DrawString(img, 10, 10, c.RandomText, color.Black)
	}

	var buf bytes.Buffer
	switch format {
	case ".jpg", ".jpeg":
		jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100}) // 图像质量值为100，是最好的图像显示
	default: // png
		png.Encode(&buf, img)
	}

	return buf.Bytes(), buf.Len()
}

// drawPixelWithColor draw pixels on img from yi, xi and randomColor with size of pixelSize x pixelSize
func drawPixelWithColor(img draw.Image, yi, xi, pixelSize int, c color.Color) {
	ys := yi * pixelSize
	ym := ys + pixelSize
	xs := xi * pixelSize
	xm := xs + pixelSize

	for y := ys; y < ym; y++ {
		for x := xs; x < xm; x++ {
			img.Set(x, y, c)
		}
	}
}

// Color generate a random color
func Color() color.Color {
	return color.RGBA{R: uint8(IntN(255)), G: uint8(IntN(255)), B: uint8(IntN(255)), A: uint8(IntN(255))}
}
