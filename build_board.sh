#!/usr/bin/env bash
git clone https://github.com/battlesnakeio/board.git build
cd build
npm install
npm run build
cp -R build ../board
cd ..
rm -rf build