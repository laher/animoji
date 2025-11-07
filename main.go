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
	var animType string

	switch subcommand {
	case "clockwise", "anticlockwise":
		animType = "rotate"
	case "hue":
		animType = "hue"
	case "zoom":
		animType = "zoom"
	case "pixelate":
		animType = "pixelate"
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
	var frames []*image.Paletted
	switch animType {
	case "rotate":
		var direction float64
		if subcommand == "clockwise" {
			direction = 1.0
		} else {
			direction = -1.0
		}
		var err error
		frames, err = generateRotateFrames(img, direction, *frameCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating frames: %v\n", err)
			os.Exit(1)
		}
	case "hue":
		var err error
		frames, err = generateHueFrames(img, *frameCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating frames: %v\n", err)
			os.Exit(1)
		}
	case "zoom":
		var err error
		frames, err = generateZoomFrames(img, *frameCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating frames: %v\n", err)
			os.Exit(1)
		}
	case "pixelate":
		var err error
		frames, err = generatePixelateFrames(img, *frameCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating frames: %v\n", err)
			os.Exit(1)
		}
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
	fmt.Fprintf(os.Stderr, "Usage: animoji <clockwise|anticlockwise|hue|zoom|pixelate> -in <input> -out <output> [flags]\n")
	fmt.Fprintf(os.Stderr, "  clockwise: Rotate image clockwise\n")
	fmt.Fprintf(os.Stderr, "  anticlockwise: Rotate image anticlockwise\n")
	fmt.Fprintf(os.Stderr, "  hue: Cycle through hue range\n")
	fmt.Fprintf(os.Stderr, "  zoom: Zoom image in (up to 6x)\n")
	fmt.Fprintf(os.Stderr, "  pixelate: Gradually pixelate image to 4x4 grid\n")
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

func generateRotateFrames(img image.Image, direction float64, frameCount int) ([]*image.Paletted, error) {
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

func generateHueFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	// Cycle through full hue range (0-360 degrees) over all frames
	hueStep := 360.0 / float64(frameCount)

	for i := 0; i < frameCount; i++ {
		hueShift := float64(i) * hueStep

		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Apply hue shift to the image
		applyHueShift(frame, img, hueShift)

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyHueShift(dst *image.RGBA, src image.Image, hueShift float64) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// Convert RGB to HSV
			h, s, v := rgbToHSV(r8, g8, b8)

			// Shift hue
			h = math.Mod(h+hueShift, 360.0)
			if h < 0 {
				h += 360.0
			}

			// Convert back to RGB
			rNew, gNew, bNew := hsvToRGB(h, s, v)

			dst.Set(x, y, color.RGBA{rNew, gNew, bNew, a8})
		}
	}
}

func rgbToHSV(r, g, b uint8) (h, s, v float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(math.Max(rf, gf), bf)
	min := math.Min(math.Min(rf, gf), bf)
	delta := max - min

	// Value
	v = max

	// Saturation
	if max == 0 {
		s = 0
	} else {
		s = delta / max
	}

	// Hue
	if delta == 0 {
		h = 0
	} else if max == rf {
		h = 60.0 * math.Mod((gf-bf)/delta+6.0, 6.0)
	} else if max == gf {
		h = 60.0 * ((bf-rf)/delta + 2.0)
	} else {
		h = 60.0 * ((rf-gf)/delta + 4.0)
	}

	if h < 0 {
		h += 360.0
	}

	return h, s, v
}

func hsvToRGB(h, s, v float64) (r, g, b uint8) {
	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2.0)-1.0))
	m := v - c

	var rf, gf, bf float64

	switch {
	case h < 60:
		rf, gf, bf = c, x, 0
	case h < 120:
		rf, gf, bf = x, c, 0
	case h < 180:
		rf, gf, bf = 0, c, x
	case h < 240:
		rf, gf, bf = 0, x, c
	case h < 300:
		rf, gf, bf = x, 0, c
	default:
		rf, gf, bf = c, 0, x
	}

	r = uint8(math.Round((rf + m) * 255.0))
	g = uint8(math.Round((gf + m) * 255.0))
	b = uint8(math.Round((bf + m) * 255.0))

	return r, g, b
}

func generateZoomFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	// Zoom from 1x to 6x over all frames
	minZoom := 1.0
	maxZoom := 6.0

	for i := 0; i < frameCount; i++ {
		// Interpolate zoom level from 1x to 6x
		progress := float64(i) / float64(frameCount-1)
		if frameCount == 1 {
			progress = 0
		}
		zoom := minZoom + (maxZoom-minZoom)*progress

		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Apply zoom to the image
		applyZoom(frame, img, zoom)

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyZoom(dst *image.RGBA, src image.Image, zoom float64) {
	dstBounds := dst.Bounds()
	dstWidth := float64(dstBounds.Dx())
	dstHeight := float64(dstBounds.Dy())

	srcBounds := src.Bounds()
	srcWidth := float64(srcBounds.Dx())
	srcHeight := float64(srcBounds.Dy())

	// Calculate the source region to sample from (centered)
	srcRegionWidth := srcWidth / zoom
	srcRegionHeight := srcHeight / zoom
	srcCenterX := srcWidth / 2.0
	srcCenterY := srcHeight / 2.0

	srcMinX := srcCenterX - srcRegionWidth/2.0
	srcMinY := srcCenterY - srcRegionHeight/2.0

	// For each pixel in destination, find corresponding pixel in source
	for y := 0; y < dstBounds.Dy(); y++ {
		for x := 0; x < dstBounds.Dx(); x++ {
			// Map destination pixel to source coordinates
			srcX := srcMinX + (float64(x)/dstWidth)*srcRegionWidth
			srcY := srcMinY + (float64(y)/dstHeight)*srcRegionHeight

			// Get pixel from source using nearest neighbor
			srcXInt := int(srcX) + srcBounds.Min.X
			srcYInt := int(srcY) + srcBounds.Min.Y

			if srcXInt >= srcBounds.Min.X && srcXInt < srcBounds.Max.X &&
				srcYInt >= srcBounds.Min.Y && srcYInt < srcBounds.Max.Y {
				dst.Set(x+dstBounds.Min.X, y+dstBounds.Min.Y, src.At(srcXInt, srcYInt))
			}
		}
	}
}

func generatePixelateFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	// Pixelate from original image (block size = 1) to 4x4 grid
	// Block size goes from 1 pixel to ensure at least 4 blocks in each dimension
	minBlockSize := 1.0
	// For 4x4 grid, we need block size that gives us at least 4 blocks
	// Use the smaller dimension to ensure we get at least 4 blocks in both directions
	maxBlockSizeX := float64(width) / 4.0
	maxBlockSizeY := float64(height) / 4.0
	maxBlockSize := math.Min(maxBlockSizeX, maxBlockSizeY)

	for i := 0; i < frameCount; i++ {
		// Interpolate block size from 1 to maxBlockSize
		progress := float64(i) / float64(frameCount-1)
		if frameCount == 1 {
			progress = 0
		}
		blockSize := minBlockSize + (maxBlockSize-minBlockSize)*progress

		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Apply pixelation to the image
		// If blockSize is 1.0, just copy the original (no pixelation)
		if blockSize <= 1.0 {
			draw.Draw(frame, frame.Bounds(), img, bounds.Min, draw.Src)
		} else {
			applyPixelate(frame, img, blockSize)
		}

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyPixelate(dst *image.RGBA, src image.Image, blockSize float64) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate number of blocks based on block size
	blocksX := int(math.Ceil(float64(width) / blockSize))
	blocksY := int(math.Ceil(float64(height) / blockSize))

	// Process each block
	for blockY := 0; blockY < blocksY; blockY++ {
		for blockX := 0; blockX < blocksX; blockX++ {
			// Calculate block boundaries
			startX := int(float64(blockX) * blockSize)
			startY := int(float64(blockY) * blockSize)
			endX := int(float64(blockX+1) * blockSize)
			endY := int(float64(blockY+1) * blockSize)

			// Clamp to bounds
			if startX < bounds.Min.X {
				startX = bounds.Min.X
			}
			if startY < bounds.Min.Y {
				startY = bounds.Min.Y
			}
			if endX > bounds.Max.X {
				endX = bounds.Max.X
			}
			if endY > bounds.Max.Y {
				endY = bounds.Max.Y
			}

			// Calculate average color of this block
			var rSum, gSum, bSum, aSum uint64
			pixelCount := 0

			for y := startY; y < endY; y++ {
				for x := startX; x < endX; x++ {
					r, g, b, a := src.At(x, y).RGBA()
					rSum += uint64(r >> 8)
					gSum += uint64(g >> 8)
					bSum += uint64(b >> 8)
					aSum += uint64(a >> 8)
					pixelCount++
				}
			}

			if pixelCount > 0 {
				avgR := uint8(rSum / uint64(pixelCount))
				avgG := uint8(gSum / uint64(pixelCount))
				avgB := uint8(bSum / uint64(pixelCount))
				avgA := uint8(aSum / uint64(pixelCount))

				// Fill the entire block with the average color
				blockColor := color.RGBA{avgR, avgG, avgB, avgA}
				for y := startY; y < endY; y++ {
					for x := startX; x < endX; x++ {
						dst.Set(x, y, blockColor)
					}
				}
			}
		}
	}
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
	for y := range height {
		for x := range width {
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
