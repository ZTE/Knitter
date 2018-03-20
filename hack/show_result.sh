#!/bin/bash
for file in $PWD/report/*
do
    if [ "${file##*.}"x = "xml"x ]||[ "${file##*.}"x = "html"x ];then
	cat $file
    fi
done
