#!/bin/bash

if [[ "$(uname)" == "Darwin"* ]]; then
	# https://www.torproject.org/download/tor/
	mkdir -p embed/mac
	echo Downloading the Tor Browser Expert Bundle for MacOS...
	curl -L -O "https://archive.torproject.org/tor-package-archive/torbrowser/13.5/tor-expert-bundle-macos-aarch64-13.5.tar.gz"
	tar xvfz tor-expert-bundle-macos-aarch64-13.5.tar.gz -C embed/mac/
	echo removing unnecessary files from the bundle...
	rm -rf embed/mac/tor/pluggable_transports/
	rm tor-expert-bundle-macos-aarch64-13.5.tar.gz
	echo the official binary is not signed, so we sign it manually...
	strip embed/mac/tor/libevent-2.1.7.dylib
	strip embed/mac/tor/tor
	codesign -s - embed/mac/tor/tor
	codesign -s - embed/mac/tor/libevent-2.1.7.dylib
	echo creating the archive...
	zip -r embed/mac/tor.zip embed/mac/tor embed/mac/data
	echo removing uncompressed files from the bundle...
	rm -rf embed/mac/data embed/mac/tor
fi
