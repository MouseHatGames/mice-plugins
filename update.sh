#!/bin/bash

pkg=$1

if [ $# == 0 ]; then
    pkg="github.com/MouseHatGames/mice"
fi

for d in */*/; do (
    echo "[*] $d"
    cd $d
    go get $pkg@latest
)
done