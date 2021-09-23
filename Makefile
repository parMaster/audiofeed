.PHONY: build
build: 
	go build -v ./main.go

.PHONY: buildstage
buildstage: 
	rm ./audiofeed
	go build -v 
	cp ./audiofeed /var/www/af.cdns.com.ua/