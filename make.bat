go build -v
@echo off
::mkdir "audio\Worlds Within Worlds. The Story of Nuclear Energy by Isaac Asimov"
::powershell -Command "Invoke-WebRequest http://www.archive.org/download/worldswithinworlds_v1-3_1904_librivox/worldswithinworlds_01_asimov_128kb.mp3 -OutFile \"audio\Worlds Within Worlds. The Story of Nuclear Energy by Isaac Asimov\worldswithinworlds_01_asimov_128kb.mp3\""
::powershell -Command "Invoke-WebRequest http://www.archive.org/download/worldswithinworlds_v1-3_1904_librivox/worldswithinworlds_02_asimov_128kb.mp3 -OutFile \"audio\Worlds Within Worlds. The Story of Nuclear Energy by Isaac Asimov\worldswithinworlds_02_asimov_128kb.mp3\""
::powershell -Command "Invoke-WebRequest https://archive.org/download/worldswithinworlds_v1-3_1904_librivox/storynuclearenergy_1904.jpg -OutFile \"audio\Worlds Within Worlds. The Story of Nuclear Energy by Isaac Asimov\storynuclearenergy_1904.jpg\""
IF %1==demo GOTO :DEMO
IF %1==run GOTO :RUN
GOTO :END

:DEMO
mkdir "audio"
IF EXIST "audio\Taras Shevchenko" (
    ECHO Demo folders might be created already
) ELSE (
    mkdir "audio\Taras Shevchenko\Kateryna"
    mkdir "audio\Taras Shevchenko\Prychynna"
    mkdir "audio\Taras Shevchenko\Zapovit"

    mkdir "audio\Alice Adventures in Wonderland abridged. Lewis Carroll"
)
IF EXIST "audio\Taras Shevchenko\Kateryna\ukrainian_kateryna_shevchenko_olga.mp3" (
    ECHO Demo files might be downloaded already
) ELSE (
    powershell -Command "Invoke-WebRequest http://www.archive.org/download/multilingual_poetry_012_0904/ukrainian_kateryna_shevchenko_olga.mp3 -OutFile \"audio\Taras Shevchenko\Kateryna\ukrainian_kateryna_shevchenko_olga.mp3\""
    powershell -Command "Invoke-WebRequest http://www.archive.org/download/multilingual_poetry_012_0904/ukrainian_prychynna_shevchenko_olga.mp3 -OutFile \"audio\Taras Shevchenko\Prychynna\ukrainian_prychynna_shevchenko_olga.mp3\""
    powershell -Command "Invoke-WebRequest http://www.archive.org/download/multilingual_short_works_collection_012_1403_librivox/msw012_20_zapovit_shevchenko_sap_128kb.mp3 -OutFile \"audio\Taras Shevchenko\Zapovit\msw012_20_zapovit_shevchenko_sap_128kb.mp3\""
    powershell -Command "Invoke-WebRequest https://archive.org/download/multilingual_short_works_collection_012_1403_librivox/multilingual_short_works_collection_012_1405.jpg -OutFile \"audio\Taras Shevchenko\multilingual_short_works_collection_012_1405.jpg\""

    powershell -Command "Invoke-WebRequest http://www.archive.org/download/alice_adventures_v_1208_librivox/alicewonderland_01_caroll.mp3 -OutFile \"audio\Alice Adventures in Wonderland abridged. Lewis Carroll\alicewonderland_01_caroll.mp3\""
    powershell -Command "Invoke-WebRequest http://archive.org/download/LibrivoxCdCoverArt19/Alices_Adventures_in_Wonderland5.jpg -OutFile \"audio\Alice Adventures in Wonderland abridged. Lewis Carroll\Alices_Adventures_in_Wonderland5.jpg\""
)
audiofeed.exe
GOTO :END

:RUN
audiofeed.exe
GOTO :END

:END