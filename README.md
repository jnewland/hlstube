# HLSTube

> On-demand HLS streams for YouTube URLs. For, uh, personal use.

## Usage

### GET /RQA5RcIZlAM/

Serve a HLS playlist, kinda like

```
mkdir -p RQA5RcIZlAM
cd RQA5RcIZlAM
youtube-dl https://www.youtube.com/watch\?v\=RQA5RcIZlAM -o - | ffmpeg -i - -c copy -f hls \
    -hls_flags omit_endlist \
    -hls_start_number_source epoch \
    -hls_wrap 10 \
    index.m3u8
```

### GET /RQA5RcIZlAM/*

Serve segments from RQA5RcIZlAM dir