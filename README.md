# Audiofeed

## ToDo
- Illustrate in screenshots the directory tree and the result on iPhone

## Demo
To build a demo app, simply run

`make demo`

Your server will be up and running after downloading some public domain content. 

Visit http://localhost:8080/index to see titles list

Demo server is also available at http://audiofeed_demo.cdns.com.ua/index 

## Usage
- Put your audiobook folder into `audio` folder
- Get RSS URL for your podcast app
- Enjoy

## Ubuntu autorun crontab line
`
@reboot cd /var/www/audiofeed_demo && ./audiofeed
`
