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

## Flags

- `-in`: Input image file (PNG or JPEG)
- `-out`: Output GIF file path
- `-frames`: Number of frames in the animation (default: 6)
- `-rate`: Frame rate in frames per second (default: 3)
- `-reverse`: Reverse the order of frames (optional)

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
```

## Animation Details

- **Rotation animation** (`360`): Rotates 360 degrees clockwise over all frames
- **Hue animation**: Cycles through full hue range (0-360 degrees) over all frames
- **Zoom animation**: Zooms from 1x to 6x over all frames
- **Pixelate animation**: Progressively pixelates from original image to 4x4 grid

The total duration of the animation is calculated as: `frames / rate` seconds.

## Requirements

- Input format: PNG or JPEG
- Output format: Animated GIF
- For rotation animation (`360`): Input image must be square
- For other animations: Any image size is supported
