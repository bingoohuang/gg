package main

import (
	"fmt"
	flag "github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/man"
	"github.com/bingoohuang/gg/pkg/randx"
)

func main() {
	w := flag.Int("w", 640, "picture width")
	h := flag.Int("h", 320, "picture height")
	fixedSizeStr := flag.String("s", "", "fixed size(eg. 44kB, 17MB)")
	many := flag.Int("i", 1, "how many pictures to create")
	picfmt := flag.String("f", "png", "picture format(png/jpg)")
	fastMode := flag.Bool("fast", false, "fast mode")
	flag.Parse()

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
