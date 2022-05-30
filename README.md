# TTS Cache

## Description

This project is a text to speech cache for [Google's Text to Speech API](https://cloud.google.com/text-to-speech).
It uses [groupcache v2](https://pkg.go.dev/github.com/mailgun/groupcache/v2)
for caching values in memory.

## Docker Testing

### Build

`docker build -t tts-cache:latest .`

### Run

```
docker run -p 8080:80 --net tts-net --env-file ./.env.0 --rm --name tts-cache-0 tts-cache
docker run -p 8081:80 --net tts-net --env-file ./.env.1 --rm --name tts-cache-1 tts-cache
docker run -p 8082:80 --net tts-net --env-file ./.env.2 --rm --name tts-cache-2 tts-cache
```

You have to create `.env.$(node)` file with the necessary environment variables
set.
