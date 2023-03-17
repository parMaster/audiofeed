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

## Usage
- Put your audiobook folder into `audio` folder
- Get RSS URL for your podcast app
- Enjoy

## Ubuntu autorun crontab line
`
@reboot cd /var/www/audiofeed_demo && ./audiofeed
`
