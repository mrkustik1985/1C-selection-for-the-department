#!/bin/bash

go build -o server ./cmd
./server
rm ./server
