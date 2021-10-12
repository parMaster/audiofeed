.PHONY: build
build: 
	go build -v

.PHONY: run
run: 
	go build -v
	./audiofeed

.PHONY: demo
demo: 
	mkdir "audio"
	mkdir "audio/Taras Shevchenko"
	mkdir "audio/Taras Shevchenko/Kateryna"
	mkdir "audio/Taras Shevchenko/Prychynna"
	mkdir "audio/Taras Shevchenko/Zapovit"
	mkdir "audio/Alice Adventures in Wonderland abridged. Lewis Carroll"
	curl http://www.archive.org/download/multilingual_poetry_012_0904/ukrainian_kateryna_shevchenko_olga.mp3 -L > audio/Taras\ Shevchenko/Kateryna/kateryna.mp3
	curl http://www.archive.org/download/multilingual_poetry_012_0904/ukrainian_prychynna_shevchenko_olga.mp3 -L > audio/Taras\ Shevchenko/Prychynna/ukrainian_prychynna_shevchenko_olga.mp3
	curl http://www.archive.org/download/multilingual_short_works_collection_012_1403_librivox/msw012_20_zapovit_shevchenko_sap_128kb.mp3 -L > audio/Taras\ Shevchenko/Zapovit/msw012_20_zapovit_shevchenko_sap_128kb.mp3
	curl https://archive.org/download/multilingual_short_works_collection_012_1403_librivox/multilingual_short_works_collection_012_1405.jpg -L > audio/Taras\ Shevchenko/multilingual_short_works_collection_012_1405.jpg
	curl http://www.archive.org/download/alice_adventures_v_1208_librivox/alicewonderland_01_caroll.mp3 -L > audio/Alice\ Adventures\ in\ Wonderland\ abridged.\ Lewis\ Carroll/alicewonderland_01_caroll.mp3
	curl http://archive.org/download/LibrivoxCdCoverArt19/Alices_Adventures_in_Wonderland5.jpg -L > audio/Alice\ Adventures\ in\ Wonderland\ abridged.\ Lewis\ Carroll/Alices_Adventures_in_Wonderland5.jpg
	./audiofeed

.PHONY: stage
stage: 
	rm ./audiofeed
	go build -v 
	cp ./audiofeed /var/www/af.cdns/