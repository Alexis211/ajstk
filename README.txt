AJSTK - A Japanese Study ToolKit

See stuff in doc/.

Configuration : copy config.json.sample to config.json and edit to your
needs.

Requirements :
- GNU/Linux or Mac OS X
- git
- Go : http://golang.org/
- godag : http://code.google.com/p/godag/
- gosqlite3 : http://github.com/kuroneko/gosqlite3

Running AJSTK with the kimidori database : you need a copy of the
kimidori-web and kimidori-contents repositories, correctly referenced in
the config.json file. You also need an empty data folder.
Building AJSTK from source : ./build.sh
Running the server : ./run.sh or ./ajstk-srv if build already done

