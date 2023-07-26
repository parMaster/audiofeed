# Audiofeed

[![Go](https://github.com/parMaster/audiofeed/actions/workflows/go.yml/badge.svg)](https://github.com/parMaster/audiofeed/actions/workflows/go.yml)

Sometimes I find downloaded audiobooks on very old HDDs, or download [free public domain audiobooks](https://librivox.org/) that I want to listen to. I find that using a podcast app is the most convenient way:
- podcast apps remember played episodes
- you can continue listen to the book from the position you stopped at the last time
- no need to download anything beforehand, listen media directly from the Internet

This app can generate XML feed that you put into your podcast app and listen to audiobook as if it were a podcast! You don't need to download media files onto your device and use any storage space.

## Demo
To build a demo app, simply run

`make demo`

Your server will be up and running after downloading some public domain content.

Visit http://localhost:8080/index to see titles list

Demo server is also available at http://audiofeed_demo.cdns.com.ua:8983/index

## Command line arguments
- `--folder=(string)` - path to folder with audiobooks (default "./audio"). Absolute or relative path allowed.
- `--code=(string)` - access code (optional) - if set, /index/{code} should be used instead of /index to access titles list, (default "")
- `--port=(int)` - port to listen to (default 8080)
- `--dbg` - enable debug mode (default false)
- `--help` - show help

## Usage
- Put your audiobook folder into `audio` folder. You can use `make demo` to download some public domain content to `audio` folder
- Run `audiofeed` with your arguments or simply `make run`
- Visit `http://localhost:8080/index` to see titles list
- Copy XML feed link and add it to your podcast app
- Enjoy

## Ubuntu autorun crontab line
`
@reboot cd /var/www/audiofeed_demo && ./audiofeed
`
