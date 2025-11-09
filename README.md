# animoji

A simple Go tool that creates animated GIFs from images with various animation effects.

## Installation

```bash
go build -o animoji .
```

## Usage

```bash
animoji [flags] <subcommand> [subcommand ...]
```

Flags come before the subcommand(s). Multiple subcommands can be chained together - effects are applied sequentially to each frame. For example:
```bash
animoji -in image.png -out output.gif -resize 128 360
animoji -in image.png -out output.gif -resize 128 ripple tint-rgb zoom
```

**Note:** This program can consume significant memory and CPU resources when processing large images. It's recommended to use the `-resize` flag to reduce image size before processing, especially for high-resolution images.

## Subcommands

### `360`
Rotates the image 360 degrees clockwise.

**Requirements:** Input image must be square (same width and height).

**Example:**
```bash
animoji -in image.png -out output.gif -resize 128 360
```

### `hue`
Cycles through the full hue range (0-360 degrees), creating a rainbow color effect.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out hue-animation.gif -resize 128 hue
```

### `zoom`
Progressively zooms into the center of the image, from 1x to 6x zoom.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out zoom-animation.gif -resize 128 zoom
```

### `pixelate`
Gradually pixelates the image, starting from the original image and ending with a 4x4 grid where each square is a uniform color.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out pixelate-animation.gif -resize 128 pixelate
```

### `tint-rgb`
Applies a tint layer with 50% opacity that cycles through RGB colors (red, yellow, green, cyan, blue, magenta) and colors in between.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out tint-animation.gif -resize 128 tint-rgb
```

### `vibes`
Divides the image into four quarters and applies rotating color tints (violet, yellow, green, blue) with 50% opacity. Each frame, the colors rotate to the next quarter.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out vibes-animation.gif -resize 128 vibes
```

### `kaleidoscope`
Creates a kaleidoscope effect with rotating mirrored sections. The image is divided into segments that are mirrored and rotated.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out kaleidoscope-animation.gif -resize 128 kaleidoscope
```

### `ripple`
Applies a ripple wave distortion that emanates from the center of the image, like dropping a stone in water.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji -in image.png -out ripple-animation.gif -resize 128 ripple
```

## Flags

- `-in`: Input image file (PNG or JPEG, optional, defaults to stdin)
- `-out`: Output GIF file path (optional, defaults to stdout)
- `-frames`: Number of frames in the animation (default: 12)
- `-rate`: Frame rate in frames per second (default: 6)
- `-reverse`: Reverse the order of frames (optional)
- `-resize`: Resize image to specified width before processing, height scaled proportionally (0 = no resize, optional)

## Examples

```bash
# Rotate 360 degrees with default settings
animoji -in square.png -out rotate.gif -resize 128 360

# Create a hue animation
animoji -in photo.jpg -out rainbow.gif -resize 128 hue

# Zoom animation with custom frame count and rate
animoji -in image.png -out zoom.gif -frames 24 -rate 10 -resize 128 zoom

# Pixelate animation with custom settings
animoji -in image.png -out pixelate.gif -frames 8 -rate 4 -resize 128 pixelate

# Zoom animation in reverse (zooms out instead of in)
animoji -in image.png -out zoom-out.gif -reverse -resize 128 zoom

# Resize image to 16 pixels wide before rotating
animoji -in large.png -out small-rotate.gif -resize 16 360

# Read from stdin and write to stdout
cat image.png | animoji -resize 128 360 > output.gif

# Read from file and write to stdout
animoji -in image.png -resize 128 360 > output.gif

# Apply RGB tint animation
animoji -in image.png -out tint.gif -resize 128 tint-rgb

# Apply vibes animation with rotating quarter tints
animoji -in image.png -out vibes.gif -resize 128 vibes

# Create kaleidoscope effect
animoji -in image.png -out kaleidoscope.gif -resize 128 kaleidoscope

# Apply ripple wave effect
animoji -in image.png -out ripple.gif -resize 128 ripple

# Chain multiple effects together (applied sequentially to each frame)
animoji -in image.png -out combined.gif -resize 128 ripple tint-rgb zoom
animoji -in image.png -out combined2.gif -resize 128 hue pixelate
```

## Animation Details

- **Rotation animation** (`360`): Rotates 360 degrees clockwise over all frames
- **Hue animation**: Cycles through full hue range (0-360 degrees) over all frames
- **Zoom animation**: Zooms from 1x to 6x over all frames
- **Pixelate animation**: Progressively pixelates from original image to 4x4 grid
- **Tint-RGB animation**: Cycles through RGB colors (red, yellow, green, cyan, blue, magenta) with 50% opacity tint layer
- **Vibes animation**: Applies rotating color tints (violet, yellow, green, blue) to image quarters with 50% opacity
- **Kaleidoscope animation**: Creates rotating mirrored segments for a kaleidoscope effect
- **Ripple animation**: Applies wave distortion emanating from the center

The total duration of the animation is calculated as: `frames / rate` seconds.

## Requirements

- Input format: PNG or JPEG
- Output format: Animated GIF
- For rotation animation (`360`): Input image must be square
- For other animations: Any image size is supported

## Performance Notes

**Warning:** This program can consume significant memory and CPU resources when processing large images. Each frame requires full image processing, and multiple effects compound the computational cost.

**Recommendations:**
- **Use the `-resize` flag to reduce image dimensions.** The resize operation occurs at the start of processing, so reducing the image size will result in much less resource usage throughout the entire animation generation process. For example, use `-resize 128` or `-resize 256` for most use cases.
- For high-resolution images (e.g., 4K or larger), always resize first to avoid excessive memory usage
- Consider reducing frame count (`-frames`) for very large images
- Processing multiple chained effects will use more resources than single effects
