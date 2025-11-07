package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	var direction float64

	switch subcommand {
	case "clockwise":
		direction = 1.0
	case "anticlockwise":
		direction = -1.0
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	fs := flag.NewFlagSet(subcommand, flag.ExitOnError)
	inFile := fs.String("in", "", "Input image file (PNG or JPEG)")
	outFile := fs.String("out", "", "Output GIF file")
	frameCount := fs.Int("frames", 6, "Number of frames in the animation")
	rate := fs.Int("rate", 3, "Frame rate in frames per second")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *inFile == "" || *outFile == "" {
		fmt.Fprintf(os.Stderr, "Both -in and -out flags are required\n")
		os.Exit(1)
	}

	if *frameCount <= 0 {
		fmt.Fprintf(os.Stderr, "Number of frames must be positive\n")
		os.Exit(1)
	}

	if *rate <= 0 {
		fmt.Fprintf(os.Stderr, "Frame rate must be positive\n")
		os.Exit(1)
	}

	// Load input image
	img, err := loadImage(*inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
		os.Exit(1)
	}

	// Generate frames
	frames, err := generateFrames(img, direction, *frameCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating frames: %v\n", err)
		os.Exit(1)
	}

	// Create animated GIF
	anim := &gif.GIF{
		Image: frames,
		Delay: make([]int, len(frames)),
	}

	// Set delay for each frame (delay in 100ths of a second)
	// delay = 100 / rate (rounded to nearest integer)
	delayPerFrame := int(math.Round(100.0 / float64(*rate)))
	for i := range anim.Delay {
		anim.Delay[i] = delayPerFrame
	}

	// Write GIF to file
	if err := writeGIF(*outFile, anim); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing GIF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created animated GIF: %s\n", *outFile)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: animoji <clockwise|anticlockwise> -in <input> -out <output> [flags]\n")
	fmt.Fprintf(os.Stderr, "  clockwise: Rotate image clockwise\n")
	fmt.Fprintf(os.Stderr, "  anticlockwise: Rotate image anticlockwise\n")
	fmt.Fprintf(os.Stderr, "  -in: Input image file (PNG or JPEG)\n")
	fmt.Fprintf(os.Stderr, "  -out: Output GIF file\n")
	fmt.Fprintf(os.Stderr, "  -frames: Number of frames in the animation (default: 6)\n")
	fmt.Fprintf(os.Stderr, "  -rate: Frame rate in frames per second (default: 3)\n")
}

func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func generateFrames(img image.Image, direction float64, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Ensure square image
	if width != height {
		return nil, fmt.Errorf("image must be square (got %dx%d)", width, height)
	}

	size := width
	center := float64(size) / 2.0

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	angleStep := (2 * math.Pi) / float64(frameCount) * direction

	for i := 0; i < frameCount; i++ {
		angle := float64(i) * angleStep

		// Create new image for this frame
		frame := image.NewRGBA(image.Rect(0, 0, size, size))

		// Rotate and draw the image
		drawRotatedImage(frame, img, center, center, angle)

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func createPalette(img image.Image) color.Palette {
	// Sample colors from the image to create a palette
	// Use a simple approach: sample every Nth pixel
	bounds := img.Bounds()
	paletteMap := make(map[color.Color]bool)
	palette := make(color.Palette, 0, 256)

	// Sample pixels
	step := 4
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			c := img.At(x, y)
			if !paletteMap[c] {
				paletteMap[c] = true
				palette = append(palette, c)
				if len(palette) >= 256 {
					break
				}
			}
		}
		if len(palette) >= 256 {
			break
		}
	}

	// Ensure we have at least some colors
	if len(palette) == 0 {
		palette = append(palette, color.White, color.Black)
	}

	return palette
}

func drawRotatedImage(dst *image.RGBA, src image.Image, cx, cy, angle float64) {
	bounds := dst.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Precompute sin and cos
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	srcBounds := src.Bounds()
	srcWidth := float64(srcBounds.Dx())
	srcHeight := float64(srcBounds.Dy())

	// For each pixel in destination, find corresponding pixel in source
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Translate to center
			dx := float64(x) - cx
			dy := float64(y) - cy

			// Rotate backwards (inverse rotation)
			sx := dx*cos + dy*sin
			sy := -dx*sin + dy*cos

			// Translate back
			sx += srcWidth / 2
			sy += srcHeight / 2

			// Get pixel from source if within bounds
			if sx >= 0 && sx < srcWidth && sy >= 0 && sy < srcHeight {
				srcX := int(sx) + srcBounds.Min.X
				srcY := int(sy) + srcBounds.Min.Y
				dst.Set(x, y, src.At(srcX, srcY))
			}
		}
	}
}

func writeGIF(filename string, anim *gif.GIF) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return gif.EncodeAll(file, anim)
}
