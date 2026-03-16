package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// ---- calcSize ---------------------------------------------------------------

func TestCalcSize_widthOnly(t *testing.T) {
	w, h := calcSize(100, 200, 500, 0)
	if w != 500 || h != 1000 {
		t.Errorf("width-only: got %dx%d, want 500x1000", w, h)
	}
}

func TestCalcSize_heightOnly(t *testing.T) {
	w, h := calcSize(100, 200, 0, 100)
	if w != 50 || h != 100 {
		t.Errorf("height-only: got %dx%d, want 50x100", w, h)
	}
}

func TestCalcSize_coverMode_widthConstrains(t *testing.T) {
	// 100x200, target 300x400 → scaleW=3.0, scaleH=2.0 → use 3.0 → 300x600
	w, h := calcSize(100, 200, 300, 400)
	if w != 300 || h != 600 {
		t.Errorf("cover (width constrains): got %dx%d, want 300x600", w, h)
	}
}

func TestCalcSize_coverMode_heightConstrains(t *testing.T) {
	// 200x100, target 300x400 → scaleW=1.5, scaleH=4.0 → use 4.0 → 800x400
	w, h := calcSize(200, 100, 300, 400)
	if w != 800 || h != 400 {
		t.Errorf("cover (height constrains): got %dx%d, want 800x400", w, h)
	}
}

func TestCalcSize_squareImage(t *testing.T) {
	w, h := calcSize(100, 100, 200, 200)
	if w != 200 || h != 200 {
		t.Errorf("square: got %dx%d, want 200x200", w, h)
	}
}

func TestCalcSize_noDownscale_widthOnly(t *testing.T) {
	// target smaller than original → still scales down
	w, h := calcSize(400, 800, 200, 0)
	if w != 200 || h != 400 {
		t.Errorf("downscale width-only: got %dx%d, want 200x400", w, h)
	}
}

func TestCalcSize_minOne(t *testing.T) {
	// extremely small target should never produce 0-pixel dimensions
	w, h := calcSize(1000, 1, 0, 1)
	if w < 1 || h < 1 {
		t.Errorf("min-one: got %dx%d, both must be >= 1", w, h)
	}
}

// ---- detectFormat -----------------------------------------------------------

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"photo.jpg", "jpeg"},
		{"photo.JPG", "jpeg"},
		{"photo.jpeg", "jpeg"},
		{"photo.JPEG", "jpeg"},
		{"photo.png", "png"},
		{"photo.PNG", "png"},
		{"photo.gif", "gif"},
		{"noextension", ""},
	}
	for _, tc := range cases {
		got := detectFormat(tc.path)
		if got != tc.want {
			t.Errorf("detectFormat(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// ---- integration tests (encode → resize → decode) ---------------------------

// newTestImage creates an in-memory RGBA image of the given size filled with a
// solid colour.
func newTestImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	return img
}

func writeJPEG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func readImageSize(t *testing.T, path string) (int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Width, cfg.Height
}

// resizeFile is the core logic extracted from main() for integration testing.
func resizeFile(t *testing.T, inputPath, outputPath string, targetW, targetH int) {
	t.Helper()

	f, err := os.Open(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()
	newW, newH := calcSize(origW, origH, targetW, targetH)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	// Use nearest-neighbor for test speed; quality is not the focus here.
	for y := range newH {
		for x := range newW {
			srcX := x * origW / newW
			srcY := y * origH / newH
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}

	out, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	outFormat := detectFormat(outputPath)
	if outFormat == "" {
		outFormat = format
	}

	switch outFormat {
	case "jpeg":
		err = jpeg.Encode(out, dst, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(out, dst)
	default:
		t.Fatalf("unsupported format: %q", outFormat)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegration_JPEG_coverMode(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.jpg")
	output := filepath.Join(dir, "out.jpg")

	writeJPEG(t, input, newTestImage(100, 200))
	resizeFile(t, input, output, 300, 400)

	w, h := readImageSize(t, output)
	if w != 300 || h != 600 {
		t.Errorf("JPEG cover: got %dx%d, want 300x600", w, h)
	}
}

func TestIntegration_PNG_coverMode(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.png")
	output := filepath.Join(dir, "out.png")

	writePNG(t, input, newTestImage(100, 200))
	resizeFile(t, input, output, 300, 400)

	w, h := readImageSize(t, output)
	if w != 300 || h != 600 {
		t.Errorf("PNG cover: got %dx%d, want 300x600", w, h)
	}
}

func TestIntegration_JPEG_widthOnly(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.jpg")
	output := filepath.Join(dir, "out.jpg")

	writeJPEG(t, input, newTestImage(100, 200))
	resizeFile(t, input, output, 500, 0)

	w, h := readImageSize(t, output)
	if w != 500 || h != 1000 {
		t.Errorf("JPEG width-only: got %dx%d, want 500x1000", w, h)
	}
}

func TestIntegration_PNG_heightOnly(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.png")
	output := filepath.Join(dir, "out.png")

	writePNG(t, input, newTestImage(100, 200))
	resizeFile(t, input, output, 0, 100)

	w, h := readImageSize(t, output)
	if w != 50 || h != 100 {
		t.Errorf("PNG height-only: got %dx%d, want 50x100", w, h)
	}
}

func TestIntegration_JPEG_inputPNG_outputJPEG(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.png")
	output := filepath.Join(dir, "out.jpg")

	writePNG(t, input, newTestImage(200, 200))
	resizeFile(t, input, output, 100, 0)

	w, h := readImageSize(t, output)
	if w != 100 || h != 100 {
		t.Errorf("PNG→JPEG: got %dx%d, want 100x100", w, h)
	}
}
