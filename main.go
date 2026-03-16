package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
)

// Version is injected at build time via -ldflags "-X main.Version=..."
var Version = "dev"

func main() {
	var (
		targetWidth  int
		targetHeight int
		outputPath   string
		showVersion  bool
	)

	flag.IntVar(&targetWidth, "width", 0, "Target width in pixels")
	flag.IntVar(&targetHeight, "height", 0, "Target height in pixels")
	flag.StringVar(&outputPath, "output", "", "Output file path (default: adds _resized suffix to input filename)")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "resize-image %s\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: resize-image [options] <input-image>\n\n")
		fmt.Fprintf(os.Stderr, "Resizes an image while maintaining aspect ratio.\n")
		fmt.Fprintf(os.Stderr, "When both --width and --height are given, the image is scaled so that\n")
		fmt.Fprintf(os.Stderr, "both dimensions are at least the specified values (cover mode).\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("resize-image %s\n", Version)
		return
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if targetWidth == 0 && targetHeight == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one of --width or --height must be specified")
		os.Exit(1)
	}
	if targetWidth < 0 || targetHeight < 0 {
		fmt.Fprintln(os.Stderr, "error: --width and --height must be positive values")
		os.Exit(1)
	}

	inputPath := args[0]

	f, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open input file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot decode image: %v\n", err)
		os.Exit(1)
	}

	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()
	newW, newH := calcSize(origW, origH, targetWidth, targetHeight)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		base := strings.TrimSuffix(inputPath, ext)
		outExt := strings.ToLower(ext)
		if outExt == ".bmp" || outExt == ".gif" {
			outExt = ".jpg"
		}
		outputPath = base + "_resized" + outExt
	}

	out, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create output file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	outputFormat := detectFormat(outputPath)
	if outputFormat == "" || outputFormat == "bmp" || outputFormat == "gif" {
		outputFormat = "jpeg"
	}

	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(out, dst, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(out, dst)
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported output format %q (supported: jpg, png)\n", outputFormat)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot encode output image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Resized: %s (%dx%d) -> %s (%dx%d)\n", inputPath, origW, origH, outputPath, newW, newH)
}

// calcSize returns new (width, height) that maintains the original aspect ratio.
// When both targetW and targetH are specified, the scale factor is chosen so that
// both dimensions are at least the target values (cover mode).
//
// Example: origW=100, origH=200, targetW=300, targetH=400
//
//	scaleW=3.0, scaleH=2.0 → use 3.0 → result 300x600
func calcSize(origW, origH, targetW, targetH int) (int, int) {
	if targetW == 0 {
		scale := float64(targetH) / float64(origH)
		return max(1, int(float64(origW)*scale)), targetH
	}
	if targetH == 0 {
		scale := float64(targetW) / float64(origW)
		return targetW, max(1, int(float64(origH)*scale))
	}

	scaleW := float64(targetW) / float64(origW)
	scaleH := float64(targetH) / float64(origH)
	scale := max(scaleW, scaleH)

	return max(1, int(float64(origW)*scale)), max(1, int(float64(origH)*scale))
}

// detectFormat returns the image format inferred from the file extension.
// Returns "" if the extension is unrecognized.
func detectFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".bmp":
		return "bmp"
	case ".gif":
		return "gif"
	default:
		return ""
	}
}
