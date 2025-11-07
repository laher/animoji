# animoji

A simple Go tool that creates animated GIFs from square images by rotating them.

## Installation

```bash
go build -o animoji .
```

## Usage

```bash
animoji <clockwise|anticlockwise> -in <input> -out <output>
```

### Options

- `clockwise`: Rotate the image clockwise
- `anticlockwise`: Rotate the image anticlockwise
- `-in`: Input image file (PNG or JPEG). Must be square.
- `-out`: Output GIF file path

### Example

```bash
animoji clockwise -in image.png -out output.gif
animoji anticlockwise -in photo.jpg -out animation.gif
```

## Animation Details

- Duration: 2 seconds
- Frames: 6 frames
- Rotation: 360 degrees total (60 degrees per frame)

## Requirements

- Input image must be square (same width and height)
- Input format: PNG or JPEG
- Output format: Animated GIF
