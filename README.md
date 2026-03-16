# resize-image

A command-line tool for resizing images while maintaining their aspect ratio.

## Overview

`resize-image` resizes a JPEG or PNG image to cover a specified target size.  
When both `--width` and `--height` are given, the image is scaled up so that
**both dimensions are at least the specified values** — not to fit inside the
target box, but to cover it (similar to CSS `object-fit: cover`).

**Example:** A 100×200 image resized with `--width 300 --height 400` becomes
**300×600** (scaled by factor 3 so width ≥ 300 and height ≥ 400).

## Installation

Download the pre-built binary for your platform from the
[Releases](https://github.com/aritelnet/resize-image/releases) page:

| Platform     | File                        |
|--------------|-----------------------------|
| Linux amd64  | `resize-image-linux-amd64`  |
| Linux arm64  | `resize-image-linux-arm64`  |

```sh
chmod +x resize-image-linux-amd64
```

## Usage

```
resize-image [options] <input-image>
```

### Options

| Flag        | Description                                                  |
|-------------|--------------------------------------------------------------|
| `--width`   | Target width in pixels                                       |
| `--height`  | Target height in pixels                                      |
| `--output`  | Output file path (default: adds `_resized` suffix to input)  |
| `--version` | Print version and exit                                       |

At least one of `--width` or `--height` must be specified.  
All flags must appear **before** the input file path.

### Examples

```sh
# Resize so width is at least 800px (height scales proportionally)
resize-image --width 800 photo.jpg

# Resize so height is at least 600px
resize-image --height 600 photo.png

# Cover 1920×1080 (both dimensions at least the target)
resize-image --width 1920 --height 1080 photo.jpg

# Specify output path
resize-image --width 300 --height 400 --output out.jpg input.jpg

# Check version
resize-image --version
```

## Supported Formats

### Input
- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- BMP (`.bmp`)
- GIF (`.gif`)

### Output
- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)

> **Note:** BMP and GIF inputs are always saved as JPEG (`.jpg`) when no `--output` path is specified.

## Building from Source

```sh
git clone https://github.com/aritelnet/resize-image.git
cd resize-image
go build -o resize-image .
```

Requires Go 1.26 or later.

## Release Process

Every push (or merge) to the `main` branch automatically:

1. Increments the patch version from the latest git tag (starting at `v0.0.1`)
2. Builds binaries for `linux/amd64` and `linux/arm64`
3. Creates a GitHub Release with the binaries attached

## License

[MIT](LICENSE)
