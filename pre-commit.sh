#!/bin/bash
ver_file=version.txt
> $ver_file
date +"%Y-%m-%d %T %:z" >> $ver_file
echo -n "Parent: " >> $ver_file
git rev-parse HEAD >> $ver_file
git add $ver_file
echo "Date-time and parent commit added to '$ver_file'"
exit 0
