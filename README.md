

WireApp Web Bot
======

This package consists of two parts:
- A golang implementation to control a WireApp account with selenium
- A command line tool based on top of that library


Installation
--
- Install `geckodriver` from https://github.com/mozilla/geckodriver/releases

```bash
# start geckodriver
geckodriver 2>&1 >> geckodriver.log &

# get this repository
go get github.com/lk16/wireapp_web_bot

# install command line tool (optional)
go install github.com/lk16/wireapp_web_bot/cmd/wireapp_web_bot
```

Usage: command line tool
--
```bash
# show help for all flags
wireapp_web_bot -h

# send message
echo "hello world" | wireapp_web_bot -user 'wireapp_user' -pass 'wireapp_password' -topic 'chat_topic' 2>>wireapp.log

# send while reading from file/pipe
cat myfile | wireapp_web_bot -user 'wireapp_user' -pass 'wireapp_password' -topic 'chat_topic' 2>>wireapp.log
```



Development details
===

TODO list:
--
- fix bug changing conversations
- fix bug with "allow notifications" pop up
- use  https://github.com/pkg/errors
- implement config file
- implement sending files
- implement interaction with sent messages
- write usage and other stuff in this file

Linter
--
```bash
gometalinter --enable-all --disable=goimports --disable=gofmt ./...
```
