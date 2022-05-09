package giff

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"os"

	_ "image/png"

	"github.com/fogleman/gg"
)

type ErrorWithInfo struct {
	Err  error
	Info string
}

func (ewi ErrorWithInfo) Error() string {
	return fmt.Sprintf("%s: %s", ewi.Info, ewi.Err.Error())
}

type Req struct {
	BgImgPath string
	FontPath  string
	FontSize  float64
	Text      string
}

func TextOnImg(req Req) (image.Image, error) {
	bgImage, err := gg.LoadPNG(req.BgImgPath)
	if err != nil {
		return nil, ErrorWithInfo{Info: "TextOnImg:gg.LoadPNG", Err: err}
	}
	imgWidth := bgImage.Bounds().Dx()
	imgHeight := bgImage.Bounds().Dy()

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(bgImage, 0, 0)

	if err := dc.LoadFontFace(req.FontPath, req.FontSize); err != nil {
		return nil, ErrorWithInfo{Info: "TextOnImg:dc.LoadFontFace", Err: err}
	}

	x := float64(imgWidth - 500)
	y := float64(imgHeight - 80)
	maxWidth := float64(imgWidth) - 60.0
	dc.SetColor(color.White)
	dc.DrawStringWrapped(req.Text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	return dc.Image(), nil
}

func CreateGIF(buf [][]byte, text []string) error {
	f, _ := os.OpenFile("earth.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	anim := gif.GIF{LoopCount: len(buf)}
	for i, v := range buf {
		if err := ioutil.WriteFile("earth.png", v, 0o644); err != nil {
			return ErrorWithInfo{Info: "CreateGIF:ioutil.WriteFile", Err: err}
		}
		img, err := TextOnImg(Req{"earth.png", "Helvetica.ttf", 100, text[i]})
		if err != nil {
			return err
		}
		bounds := img.Bounds()
		drawer := draw.FloydSteinberg
		palettedImg := image.NewPaletted(bounds, palette.Plan9)
		drawer.Draw(palettedImg, bounds, img, image.ZP)
		anim.Image = append(anim.Image, palettedImg)
		anim.Delay = append(anim.Delay, 50)
	}
	encodeErr := gif.EncodeAll(f, &anim)
	if encodeErr != nil {
		return ErrorWithInfo{Info: "CreateGIF:gif.EncodeAll", Err: encodeErr}
	}
	return nil
}
