# Audiofeed

## ToDo
- Description about Golang version
- Golang binary installation manual
- Illustrate in screenshots the directory tree and the result on iPhone

## Demo
- Visit http://audiofeed_demo.cdns.com.ua/index/ to check out the demo instance of the service with a couple of public-domain audiobooks

## Installation
- Create folder "audio"
- Put some audiobooks - each in separate folder
- edit config.php - HTTP_ADDRESS is enough
- edit .htaccess - all the books.mydomain.com entries

## Usage
- Visit http://books.mydomain.com/index (go to your actual HTTP_ADDRESS)
- Copy URL of any book
- Put the url into Podcasts App
- Enjoy

## Ubuntu autorun crontab line
`
@reboot cd /var/www/audiofeed_demo && ./audiofeed
`
