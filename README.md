# HLSTube

> On-demand HLS streams using youtube-dl. For, uh, personal use.

## Project status

[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://www.repostatus.org/badges/latest/wip.svg)](https://www.repostatus.org/#wip)

### Known issues

- Supports YouTube livestreams for which a combined video+audio format is available
- Occasional stream disconnects
- Occasional stream initialization errors

## Install

[Pull, clone, or download the latest release](https://github.com/jnewland/hlstube/releases/latest).

## Usage

```
docker run --rm -i -p 8080:8080 ghcr.io/jnewland/hlstube:latest
```

* watch a video: http://localhost:8080/_/http://youtube.com/watch?v=RQA5RcIZlAM
* watch a video: http://localhost:8080/RQA5RcIZlAM
* redirect to a video: http://localhost:8080/r/_/RQA5RcIZlAM

## Development

* Codespaces
* Cmd-shift-b or Run build task