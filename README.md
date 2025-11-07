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

### `clockwise`
Rotates the image clockwise through 360 degrees.

**Requirements:** Input image must be square (same width and height).

**Example:**
```bash
animoji clockwise -in image.png -out output.gif -frames 12 -rate 6
```

### `anticlockwise`
Rotates the image anticlockwise (counter-clockwise) through 360 degrees.

**Requirements:** Input image must be square (same width and height).

**Example:**
```bash
animoji anticlockwise -in photo.jpg -out animation.gif
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

## Examples

```bash
# Rotate clockwise with default settings (6 frames, 3 fps)
animoji clockwise -in square.png -out rotate.gif

# Create a hue animation with 12 frames at 6 fps
animoji hue -in photo.jpg -out rainbow.gif -frames 12 -rate 6

# Zoom animation with custom frame count
animoji zoom -in image.png -out zoom.gif -frames 24 -rate 10

# Pixelate animation
animoji pixelate -in image.png -out pixelate.gif -frames 8 -rate 4
```

## Animation Details

- **Rotation animations** (`clockwise`, `anticlockwise`): Rotate 360 degrees total over all frames
- **Hue animation**: Cycles through full hue range (0-360 degrees) over all frames
- **Zoom animation**: Zooms from 1x to 6x over all frames
- **Pixelate animation**: Progressively pixelates from original image to 4x4 grid

The total duration of the animation is calculated as: `frames / rate` seconds.

## Requirements

- Input format: PNG or JPEG
- Output format: Animated GIF
- For rotation animations (`clockwise`, `anticlockwise`): Input image must be square
- For other animations: Any image size is supported
