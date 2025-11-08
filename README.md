# animoji

A simple Go tool that creates animated GIFs from images with various animation effects.

## Installation

```bash
go build -o animoji .
```

## Usage

```bash
animoji <subcommand> -in <input> -out <output> [flags]
```

## Subcommands

### `360`
Rotates the image 360 degrees clockwise.

**Requirements:** Input image must be square (same width and height).

**Example:**
```bash
animoji 360 -in image.png -out output.gif -frames 12 -rate 6
```

### `hue`
Cycles through the full hue range (0-360 degrees), creating a rainbow color effect.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji hue -in image.png -out hue-animation.gif -frames 12 -rate 6
```

### `zoom`
Progressively zooms into the center of the image, from 1x to 6x zoom.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji zoom -in image.png -out zoom-animation.gif -frames 12 -rate 6
```

### `pixelate`
Gradually pixelates the image, starting from the original image and ending with a 4x4 grid where each square is a uniform color.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji pixelate -in image.png -out pixelate-animation.gif -frames 12 -rate 6
```

### `tint-rgb`
Applies a tint layer with 50% opacity that cycles through RGB colors (red, yellow, green, cyan, blue, magenta) and colors in between.

**Requirements:** Works with any image size.

**Example:**
```bash
animoji tint-rgb -in image.png -out tint-animation.gif -frames 12 -rate 6
```

## Flags

- `-in`: Input image file (PNG or JPEG, optional, defaults to stdin)
- `-out`: Output GIF file path (optional, defaults to stdout)
- `-frames`: Number of frames in the animation (default: 6)
- `-rate`: Frame rate in frames per second (default: 3)
- `-reverse`: Reverse the order of frames (optional)
- `-resize`: Resize image to specified width before processing, height scaled proportionally (0 = no resize, optional)

## Examples

```bash
# Rotate 360 degrees with default settings (6 frames, 3 fps)
animoji 360 -in square.png -out rotate.gif

# Create a hue animation with 12 frames at 6 fps
animoji hue -in photo.jpg -out rainbow.gif -frames 12 -rate 6

# Zoom animation with custom frame count
animoji zoom -in image.png -out zoom.gif -frames 24 -rate 10

# Pixelate animation
animoji pixelate -in image.png -out pixelate.gif -frames 8 -rate 4

# Zoom animation in reverse (zooms out instead of in)
animoji zoom -in image.png -out zoom-out.gif -frames 12 -rate 6 -reverse

# Resize image to 16 pixels wide before rotating
animoji 360 -in large.png -out small-rotate.gif -resize 16

# Read from stdin and write to stdout
cat image.png | animoji 360 -frames 12 -rate 6 > output.gif

# Read from file and write to stdout
animoji 360 -in image.png -frames 12 -rate 6 > output.gif

# Apply RGB tint animation
animoji tint-rgb -in image.png -out tint.gif -frames 12 -rate 6
```

## Animation Details

- **Rotation animation** (`360`): Rotates 360 degrees clockwise over all frames
- **Hue animation**: Cycles through full hue range (0-360 degrees) over all frames
- **Zoom animation**: Zooms from 1x to 6x over all frames
- **Pixelate animation**: Progressively pixelates from original image to 4x4 grid
- **Tint-RGB animation**: Cycles through RGB colors (red, yellow, green, cyan, blue, magenta) with 50% opacity tint layer

The total duration of the animation is calculated as: `frames / rate` seconds.

## Requirements

- Input format: PNG or JPEG
- Output format: Animated GIF
- For rotation animation (`360`): Input image must be square
- For other animations: Any image size is supported
