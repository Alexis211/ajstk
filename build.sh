#!/bin/sh

cd $(dirname $0)

gd -o ajstk-srv src

#8g -I src/ -o src/util.8\
#	src/util/util.go

#8g -I src/ -o src/contents.8\
#	src/contents/data.go\
#	src/contents/parser.go

#8g -I src/ -o src/study.8\
#	src/study/study.go

#8g -I src/ -o src/webs.8\
#	src/webs/webs.go\
#	src/webs/views.go\
#	src/webs/session.go\
#	src/webs/util.go

#8g -I src/ -o src/main.8\
#	src/main/main.go\
#	src/main/config.go

#8l -L src/ -o ajstk-srv src/main.8

echo "Build done. Executable is ./ajstk-srv"
