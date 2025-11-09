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
	"io"
	"math"
	"os"
)

func main() {
	inFile := flag.String("in", "", "Input image file (PNG or JPEG)")
	outFile := flag.String("out", "", "Output GIF file")
	frameCount := flag.Int("frames", 12, "Number of frames in the animation")
	rate := flag.Int("rate", 6, "Frame rate in frames per second")
	reverse := flag.Bool("reverse", false, "Reverse the order of frames")
	resize := flag.Int("resize", 0, "Resize image to specified width (height scaled proportionally, 0 = no resize)")

	flag.Parse()

	// Get subcommands from remaining arguments
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: at least one subcommand is required\n")
		printUsage()
		os.Exit(1)
	}

	// Validate all subcommands
	validSubcommands := map[string]bool{
		"360":          true,
		"hue":          true,
		"zoom":          true,
		"pixelate":      true,
		"tint-rgb":     true,
		"vibes":         true,
		"kaleidoscope": true,
		"ripple":        true,
	}

	subcommands := args
	for _, subcommand := range subcommands {
		if !validSubcommands[subcommand] {
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
			printUsage()
			os.Exit(1)
		}
	}

	if *frameCount <= 0 {
		fmt.Fprintf(os.Stderr, "Number of frames must be positive\n")
		os.Exit(1)
	}

	if *rate <= 0 {
		fmt.Fprintf(os.Stderr, "Frame rate must be positive\n")
		os.Exit(1)
	}

	if *resize < 0 {
		fmt.Fprintf(os.Stderr, "Resize width must be non-negative\n")
		os.Exit(1)
	}

	// Load input image
	var img image.Image
	var err error
	if *inFile == "" {
		img, err = loadImageFromReader(os.Stdin)
	} else {
		img, err = loadImage(*inFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
		os.Exit(1)
	}

	// Resize image if requested
	if *resize > 0 {
		img, err = resizeImage(img, *resize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resizing image: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate frames by applying all effects sequentially to each frame
	frames := make([]*image.Paletted, *frameCount)
	palette := createPalette(img)

	for i := 0; i < *frameCount; i++ {
		// Start with the original image
		currentImg := img

		// Apply each effect in sequence
		for _, subcommand := range subcommands {
			var err error
			currentImg, err = applyEffectToFrame(currentImg, subcommand, i, *frameCount)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error applying effect %s to frame %d: %v\n", subcommand, i, err)
				os.Exit(1)
			}
		}

		// Convert to paletted image for GIF
		rgba := image.NewRGBA(currentImg.Bounds())
		draw.Draw(rgba, rgba.Bounds(), currentImg, currentImg.Bounds().Min, draw.Src)

		paletted := image.NewPaletted(rgba.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), rgba, rgba.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	// Reverse frames if requested
	if *reverse {
		for i, j := 0, len(frames)-1; i < j; i, j = i+1, j-1 {
			frames[i], frames[j] = frames[j], frames[i]
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

	// Write GIF to file or stdout
	if *outFile == "" {
		if err := writeGIFToWriter(os.Stdout, anim); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing GIF: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := writeGIF(*outFile, anim); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing GIF: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully created animated GIF: %s\n", *outFile)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: animoji [flags] <subcommand>\n")
	fmt.Fprintf(os.Stderr, "\nFlags:\n")
	fmt.Fprintf(os.Stderr, "  -in: Input image file (PNG or JPEG, optional, defaults to stdin)\n")
	fmt.Fprintf(os.Stderr, "  -out: Output GIF file (optional, defaults to stdout)\n")
	fmt.Fprintf(os.Stderr, "  -frames: Number of frames in the animation (default: 6)\n")
	fmt.Fprintf(os.Stderr, "  -rate: Frame rate in frames per second (default: 3)\n")
	fmt.Fprintf(os.Stderr, "  -reverse: Reverse the order of frames (optional)\n")
	fmt.Fprintf(os.Stderr, "  -resize: Resize image to specified width, height scaled proportionally (0 = no resize)\n")
	fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
	fmt.Fprintf(os.Stderr, "  360: Rotate image 360 degrees clockwise\n")
	fmt.Fprintf(os.Stderr, "  hue: Cycle through hue range\n")
	fmt.Fprintf(os.Stderr, "  zoom: Zoom image in (up to 6x)\n")
	fmt.Fprintf(os.Stderr, "  pixelate: Gradually pixelate image to 4x4 grid\n")
	fmt.Fprintf(os.Stderr, "  tint-rgb: Apply RGB tint layer with 50%% opacity, cycling through colors\n")
	fmt.Fprintf(os.Stderr, "  vibes: Apply rotating color tints to image quarters (violet, yellow, green, blue)\n")
	fmt.Fprintf(os.Stderr, "  kaleidoscope: Create kaleidoscope effect with rotating mirrored sections\n")
	fmt.Fprintf(os.Stderr, "  ripple: Apply ripple wave distortion emanating from center\n")
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  animoji -in image.png -out output.gif -resize 128 360\n")
	fmt.Fprintf(os.Stderr, "  animoji -in image.png -out output.gif -resize 128 ripple tint-rgb zoom\n")
	fmt.Fprintf(os.Stderr, "\nNote: Multiple subcommands can be chained together. Effects are applied sequentially to each frame.\n")
}

func applyEffectToFrame(img image.Image, subcommand string, frameIdx, frameCount int) (image.Image, error) {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	switch subcommand {
	case "360":
		direction := 1.0
		angle := float64(frameIdx) * 2.0 * math.Pi / float64(frameCount) * direction
		width := bounds.Dx()
		height := bounds.Dy()
		if width != height {
			return nil, fmt.Errorf("image must be square (got %dx%d)", width, height)
		}
		size := width
		center := float64(size) / 2.0
		drawRotatedImage(result, img, center, center, angle)
		return result, nil

	case "hue":
		hueShift := float64(frameIdx) * 360.0 / float64(frameCount)
		applyHueShift(result, img, hueShift)
		return result, nil

	case "zoom":
		minZoom := 1.0
		maxZoom := 6.0
		progress := float64(frameIdx) / float64(frameCount-1)
		if frameCount == 1 {
			progress = 0
		}
		zoom := minZoom + (maxZoom-minZoom)*progress
		if zoom <= 1.0 {
			draw.Draw(result, result.Bounds(), img, bounds.Min, draw.Src)
		} else {
			applyZoom(result, img, zoom)
		}
		return result, nil

	case "pixelate":
		width := bounds.Dx()
		height := bounds.Dy()
		minBlockSize := 1.0
		maxBlockSizeX := float64(width) / 4.0
		maxBlockSizeY := float64(height) / 4.0
		maxBlockSize := math.Min(maxBlockSizeX, maxBlockSizeY)
		progress := float64(frameIdx) / float64(frameCount-1)
		if frameCount == 1 {
			progress = 0
		}
		blockSize := minBlockSize + (maxBlockSize-minBlockSize)*progress
		if blockSize <= 1.0 {
			draw.Draw(result, result.Bounds(), img, bounds.Min, draw.Src)
		} else {
			applyPixelate(result, img, blockSize)
		}
		return result, nil

	case "tint-rgb":
		hue := float64(frameIdx) * 360.0 / float64(frameCount)
		applyTint(result, img, hue)
		return result, nil

	case "vibes":
		width := bounds.Dx()
		height := bounds.Dy()
		colors := []color.RGBA{
			{255, 20, 147, 255},  // Hot Pink/Magenta
			{255, 255, 0, 255},   // Bright Yellow
			{50, 255, 50, 255},    // Bright Lime Green
			{0, 200, 255, 255},   // Bright Cyan Blue
		}
		// Draw base image first
		draw.Draw(result, result.Bounds(), img, bounds.Min, draw.Src)
		// Apply tints to quarters
		for quarter := 0; quarter < 4; quarter++ {
			colorIndex := (frameIdx + quarter) % 4
			tintColor := colors[colorIndex]
			var startX, endX, startY, endY int
			switch quarter {
			case 0: // Top-left
				startX = bounds.Min.X
				endX = bounds.Min.X + width/2
				startY = bounds.Min.Y
				endY = bounds.Min.Y + height/2
			case 1: // Top-right
				startX = bounds.Min.X + width/2
				endX = bounds.Max.X
				startY = bounds.Min.Y
				endY = bounds.Min.Y + height/2
			case 2: // Bottom-left
				startX = bounds.Min.X
				endX = bounds.Min.X + width/2
				startY = bounds.Min.Y + height/2
				endY = bounds.Max.Y
			case 3: // Bottom-right
				startX = bounds.Min.X + width/2
				endX = bounds.Max.X
				startY = bounds.Min.Y + height/2
				endY = bounds.Max.Y
			}
			applyTintToRegion(result, img, tintColor, startX, endX, startY, endY)
		}
		return result, nil

	case "kaleidoscope":
		width := bounds.Dx()
		height := bounds.Dy()
		centerX := float64(width) / 2.0
		centerY := float64(height) / 2.0
		rotationAngle := float64(frameIdx) * 2.0 * math.Pi / float64(frameCount)
		applyKaleidoscope(result, img, centerX, centerY, rotationAngle)
		return result, nil

	case "ripple":
		width := bounds.Dx()
		height := bounds.Dy()
		centerX := float64(width) / 2.0
		centerY := float64(height) / 2.0
		maxDistance := math.Sqrt(centerX*centerX + centerY*centerY)
		phase := float64(frameIdx) * 2.0 * math.Pi / float64(frameCount)
		applyRipple(result, img, centerX, centerY, phase, maxDistance)
		return result, nil

	default:
		return nil, fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return loadImageFromReader(file)
}

func loadImageFromReader(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func resizeImage(img image.Image, targetWidth int) (image.Image, error) {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth == 0 || srcHeight == 0 {
		return nil, fmt.Errorf("source image has zero dimensions")
	}

	// Calculate target height maintaining aspect ratio
	targetHeight := int(float64(targetWidth) * float64(srcHeight) / float64(srcWidth))

	// Create new RGBA image with target dimensions
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Scale factors
	scaleX := float64(srcWidth) / float64(targetWidth)
	scaleY := float64(srcHeight) / float64(targetHeight)

	// Resize using nearest neighbor sampling
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			// Map destination pixel to source coordinates
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)

			// Clamp to source bounds
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}

			// Get pixel from source
			dst.Set(x, y, img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
		}
	}

	return dst, nil
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

func generateTintRGBFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	// Cycle through RGB colors: Red -> Yellow -> Green -> Cyan -> Blue -> Magenta -> Red
	// This is essentially cycling through hue 0-360 degrees

	for i := 0; i < frameCount; i++ {
		// Calculate hue for this frame (0-360 degrees)
		hue := float64(i) * 360.0 / float64(frameCount)

		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Apply tint to the image
		applyTint(frame, img, hue)

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyTint(dst *image.RGBA, src image.Image, hue float64) {
	bounds := src.Bounds()
	opacity := 0.5 // 50% opacity

	// Convert hue to RGB color
	r, g, b := hsvToRGB(hue, 1.0, 1.0)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get source pixel
			srcR, srcG, srcB, srcA := src.At(x, y).RGBA()
			srcR8 := uint8(srcR >> 8)
			srcG8 := uint8(srcG >> 8)
			srcB8 := uint8(srcB >> 8)
			srcA8 := uint8(srcA >> 8)

			// Blend tint color with source pixel at 50% opacity
			// Formula: result = source * (1 - opacity) + tint * opacity
			blendR := uint8(float64(srcR8)*(1.0-opacity) + float64(r)*opacity)
			blendG := uint8(float64(srcG8)*(1.0-opacity) + float64(g)*opacity)
			blendB := uint8(float64(srcB8)*(1.0-opacity) + float64(b)*opacity)

			dst.Set(x, y, color.RGBA{blendR, blendG, blendB, srcA8})
		}
	}
}

func generateVibesFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create palette from source image
	palette := createPalette(img)

	// Define the four colors: violet, yellow, green, blue
	// Using vibrant highlighter pen colors
	colors := []color.RGBA{
		{255, 20, 147, 255}, // Hot Pink/Magenta (vibrant violet/pink highlighter)
		{255, 255, 0, 255},  // Bright Yellow (classic yellow highlighter)
		{50, 255, 50, 255},  // Bright Lime Green (vibrant green highlighter)
		{0, 200, 255, 255},  // Bright Cyan Blue (vibrant blue highlighter)
	}

	frames := make([]*image.Paletted, frameCount)

	for i := 0; i < frameCount; i++ {
		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Calculate which color each quarter should have
		// Each frame, rotate the colors: quarter 0 gets color (i+0)%4, quarter 1 gets (i+1)%4, etc.
		for quarter := 0; quarter < 4; quarter++ {
			colorIndex := (i + quarter) % 4
			tintColor := colors[colorIndex]

			// Calculate quarter boundaries
			var startX, endX, startY, endY int
			switch quarter {
			case 0: // Top-left
				startX = bounds.Min.X
				endX = bounds.Min.X + width/2
				startY = bounds.Min.Y
				endY = bounds.Min.Y + height/2
			case 1: // Top-right
				startX = bounds.Min.X + width/2
				endX = bounds.Max.X
				startY = bounds.Min.Y
				endY = bounds.Min.Y + height/2
			case 2: // Bottom-left
				startX = bounds.Min.X
				endX = bounds.Min.X + width/2
				startY = bounds.Min.Y + height/2
				endY = bounds.Max.Y
			case 3: // Bottom-right
				startX = bounds.Min.X + width/2
				endX = bounds.Max.X
				startY = bounds.Min.Y + height/2
				endY = bounds.Max.Y
			}

			// Apply tint to this quarter
			applyTintToRegion(frame, img, tintColor, startX, endX, startY, endY)
		}

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyTintToRegion(dst *image.RGBA, src image.Image, tintColor color.RGBA, startX, endX, startY, endY int) {
	opacity := 0.5 // 50% opacity
	tintR := float64(tintColor.R)
	tintG := float64(tintColor.G)
	tintB := float64(tintColor.B)

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			// Get source pixel
			srcR, srcG, srcB, srcA := src.At(x, y).RGBA()
			srcR8 := uint8(srcR >> 8)
			srcG8 := uint8(srcG >> 8)
			srcB8 := uint8(srcB >> 8)
			srcA8 := uint8(srcA >> 8)

			// Blend tint color with source pixel at 50% opacity
			blendR := uint8(float64(srcR8)*(1.0-opacity) + tintR*opacity)
			blendG := uint8(float64(srcG8)*(1.0-opacity) + tintG*opacity)
			blendB := uint8(float64(srcB8)*(1.0-opacity) + tintB*opacity)

			dst.Set(x, y, color.RGBA{blendR, blendG, blendB, srcA8})
		}
	}
}

func generateKaleidoscopeFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0

	for i := 0; i < frameCount; i++ {
		// Rotate the kaleidoscope pattern
		rotationAngle := float64(i) * 2.0 * math.Pi / float64(frameCount)

		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Apply kaleidoscope effect
		applyKaleidoscope(frame, img, centerX, centerY, rotationAngle)

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyKaleidoscope(dst *image.RGBA, src image.Image, cx, cy, rotationAngle float64) {
	bounds := dst.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Number of segments (like a kaleidoscope mirror)
	segments := 8

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Translate to center
			dx := float64(x) - cx
			dy := float64(y) - cy

			// Calculate angle and distance from center
			angle := math.Atan2(dy, dx) + rotationAngle
			distance := math.Sqrt(dx*dx + dy*dy)

			// Map to one segment (0 to 2*pi/segments)
			segmentAngle := math.Mod(angle, 2.0*math.Pi/float64(segments))
			if segmentAngle < 0 {
				segmentAngle += 2.0 * math.Pi / float64(segments)
			}

			// Mirror within the segment
			if segmentAngle > math.Pi/float64(segments) {
				segmentAngle = 2.0*math.Pi/float64(segments) - segmentAngle
			}

			// Calculate source coordinates
			srcAngle := segmentAngle - rotationAngle
			srcX := int(cx + distance*math.Cos(srcAngle))
			srcY := int(cy + distance*math.Sin(srcAngle))

			// Clamp to source bounds
			if srcX >= bounds.Min.X && srcX < bounds.Max.X &&
				srcY >= bounds.Min.Y && srcY < bounds.Max.Y {
				dst.Set(x, y, src.At(srcX, srcY))
			}
		}
	}
}

func generateRippleFrames(img image.Image, frameCount int) ([]*image.Paletted, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create palette from source image
	palette := createPalette(img)

	frames := make([]*image.Paletted, frameCount)
	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0
	maxDistance := math.Sqrt(centerX*centerX + centerY*centerY)

	for i := 0; i < frameCount; i++ {
		// Ripple phase (0 to 2*pi)
		phase := float64(i) * 2.0 * math.Pi / float64(frameCount)

		// Create new image for this frame
		frame := image.NewRGBA(bounds)

		// Apply ripple effect
		applyRipple(frame, img, centerX, centerY, phase, maxDistance)

		// Convert to paletted image for GIF
		paletted := image.NewPaletted(frame.Bounds(), palette)
		draw.Draw(paletted, paletted.Bounds(), frame, frame.Bounds().Min, draw.Src)

		frames[i] = paletted
	}

	return frames, nil
}

func applyRipple(dst *image.RGBA, src image.Image, cx, cy, phase, maxDistance float64) {
	bounds := dst.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Ripple parameters
	amplitude := 5.0  // Maximum pixel displacement
	frequency := 0.1  // Ripple frequency

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate distance from center
			dx := float64(x) - cx
			dy := float64(y) - cy
			distance := math.Sqrt(dx*dx + dy*dy)

			// Calculate ripple displacement
			ripple := amplitude * math.Sin(distance*frequency-phase)

			// Calculate angle
			angle := math.Atan2(dy, dx)

			// Apply displacement along the radial direction
			displacedDistance := distance + ripple
			srcX := int(cx + displacedDistance*math.Cos(angle))
			srcY := int(cy + displacedDistance*math.Sin(angle))

			// Clamp to source bounds
			if srcX >= bounds.Min.X && srcX < bounds.Max.X &&
				srcY >= bounds.Min.Y && srcY < bounds.Max.Y {
				dst.Set(x, y, src.At(srcX, srcY))
			} else {
				// If out of bounds, use nearest edge pixel
				srcX = int(math.Max(float64(bounds.Min.X), math.Min(float64(bounds.Max.X-1), float64(srcX))))
				srcY = int(math.Max(float64(bounds.Min.Y), math.Min(float64(bounds.Max.Y-1), float64(srcY))))
				dst.Set(x, y, src.At(srcX, srcY))
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

	return writeGIFToWriter(file, anim)
}

func writeGIFToWriter(w io.Writer, anim *gif.GIF) error {
	return gif.EncodeAll(w, anim)
}
