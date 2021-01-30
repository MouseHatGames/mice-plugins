#!/bin/bash

pkg=$1

if [ $# == 0 ]; then
    pkg="github.com/MouseHatGames/mice"
fi

for d in */*/; do (
    echo "[*] $d"
    cd $d
    GO111MODULE=on go get $pkg@latest
)
done