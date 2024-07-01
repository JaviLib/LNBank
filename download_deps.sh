#!/bin/bash

# TODO check shasum
if [[ "$(uname)" == "Darwin"* ]]; then
	# https://www.torproject.org/download/tor/
	mkdir -p embed
	echo Downloading the Tor Browser Expert Bundle for MacOS...
	curl -L -O "https://archive.torproject.org/tor-package-archive/torbrowser/13.5/tor-expert-bundle-macos-aarch64-13.5.tar.gz"
	tar xvfz tor-expert-bundle-macos-aarch64-13.5.tar.gz -C embed
	echo removing unnecessary files from the bundle...
	rm -rf embed/tor/pluggable_transports/
	rm tor-expert-bundle-macos-aarch64-13.5.tar.gz
	echo the official binary is not signed, so we sign it manually...
	strip embed/tor/libevent-2.1.7.dylib
	strip embed/tor/tor
	codesign -s - embed/tor/tor
	codesign -s - embed/tor/libevent-2.1.7.dylib
	echo creating the archive...
	zip -9 -r embed/tor.zip embed/tor embed/data
	echo removing uncompressed files from the bundle...
	rm -rf embed/data embed/tor

	# https://github.com/lightningnetwork/lnd/releases
	echo Downloading lnd for MacOS
	curl -L -O "https://github.com/lightningnetwork/lnd/releases/download/v0.18.1-beta/lnd-darwin-arm64-v0.18.1-beta.tar.gz"
	tar xvfz lnd-darwin-arm64-v0.18.1-beta.tar.gz -C embed
	mv embed/lnd-darwin-arm64-v0.18.1-beta embed/lnd
	rm lnd-darwin-arm64-v0.18.1-beta.tar.gz
	echo the official binary is not signed, so we sign it manually...
	strip embed/lnd/lnd
	strip embed/lnd/lncli
	codesign -s - embed/lnd/*
	echo creating the archive...
	zip -9 -r embed/lnd.zip embed/lnd
	rm -rf embed/lnd
fi
