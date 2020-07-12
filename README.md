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

### Deferred image decoding

Image files may be read in without decoding the actual image data, and
these deferred images may be written out without incurring the expense
and potential image loss of decoding or encoding them. This is useful
for code which only wants to examine image metadata, but it also
allows programs that want to mutate an image's metadata without
mutating the image itself to do so inexpensively and without altering
image quality.

This is especially useful for lossy image formats such as jpeg --
deferred image reading and writing allows programs to update jpeg
metadata without having to re-encode the image.