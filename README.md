# HLSTube

[![Project Status: Unsupported – The project has reached a stable, usable state but the author(s) have ceased all work on it. A new maintainer may be desired.](https://www.repostatus.org/badges/latest/unsupported.svg)](https://www.repostatus.org/#unsupported)

## Project status

[![Project Status: Unsupported – The project has reached a stable, usable state but the author(s) have ceased all work on it. A new maintainer may be desired.](https://www.repostatus.org/badges/latest/unsupported.svg)](https://www.repostatus.org/#unsupported)

### Known issues

- Supports YouTube livestreams for which a combined video+audio format is available
- Occasional stream disconnects
- Occasional stream initialization errors

## Install

[Pull](https://ghcr.io/jnewland/hlstube), clone, or download the [latest release](https://github.com/jnewland/hlstube/releases/latest).

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
