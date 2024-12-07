@echo off
cd /d "C:\Program Files\JACK2"
.\jackd.exe -S -X winmme -dportaudio -d"ASIO::ASIO4ALL v2" -r48000 -p128
pause
