# third_party directory

The `third_party` dir contains modified packages or a code snippets taken from internet.

## dca package - Go implementation for the DCA audio format
This version was modified:
- supports more parameters passed to FFMPEG
- replaced existed logger to `slog` for consistency with Melodix logging system.

Based on forked repo [ClintonCollins GitHub](https://github.com/ClintonCollins/dca).
Original repo [jonas747 GitHub repo](https://github.com/jonas747/dca)