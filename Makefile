.PHONY: build
build: 
	go build -v

.PHONY: stage
stage: 
	rm ./audiofeed
	go build -v 
	cp ./audiofeed /var/www/af.cdns.com.ua/