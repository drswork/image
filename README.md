# image

Image, an updated version of the Go standard image libraries

## Features

### Metadata support

This package supports reading and writing image metadata. This ranges
from things as simple as comment blocks in GIF files to ICC color
profiles and XMP data.

### Size and time safety

Code can set limits on the amount of wall time or memory that encoding
or decoding an image can use, which allows code to read files with
more safety guarantees than the standard library.

### Updated API

A new, unified API exists that allows access to the image metadata,
control over encoding and decoding, and the writing of image files
with metadata.

The existing API has been maintained for drop-in compatibility with
the current standard library.