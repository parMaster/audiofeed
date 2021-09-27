.PHONY: build
build: 
	go build -v

.PHONY: buildstage
buildstage: 
	rm ./audiofeed
	go build -v 
	cp ./audiofeed /var/www/af.cdns.com.ua/