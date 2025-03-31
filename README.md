## HLS Converter

### Install FFMPEG to host

```zsh
sudo apt-get update
sudo apt install ffmpeg
```

### Put FFMPEG Library Volume into container

```yml
        # 호스트의 FFmpeg 바이너리 마운트
      - /usr/bin/ffmpeg:/usr/bin/ffmpeg:ro
      # FFmpeg 관련 공유 라이브러리 마운트 (Ubuntu/Debian 기준)
      - /usr/lib:/usr/lib:ro
      - /lib:/lib:ro
```
