package main

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/man"
	"github.com/bingoohuang/gg/pkg/randx"
)

func main() {
	w := fla9.Int("width,w", 640, "picture width")
	h := fla9.Int("height,h", 320, "picture height")
	fixedSizeStr := fla9.String("size,s", "", "fixed size(eg. 44kB, 17MB)")
	many := fla9.Int("n", 1, "how many pictures to create")
	picfmt := fla9.String("format,f", "png", "picture format(png/jpg)")
	fastMode := fla9.Bool("fast", false, "fast mode")
	fla9.Parse()

	fixedSize := uint64(0)
	if *fixedSizeStr != "" {
		var err error
		if fixedSize, err = man.ParseBytes(*fixedSizeStr); err != nil {
			panic("illegal fixed size " + err.Error())
		}
	}

	s := randx.Int()

	for i := 0; i < *many; i++ {
		randTxt := fmt.Sprintf("%d", s+i)
		fn := fmt.Sprintf("%d_%dx%d.%s", s+i, *w, *h, *picfmt)
		c := randx.ImgConfig{Width: *w, Height: *h, RandomText: randTxt, FastMode: *fastMode}
		size := c.GenFile(fn, int(fixedSize))

		fmt.Println(fn, "size", man.IBytes(uint64(size)), "generated!")
	}
}
