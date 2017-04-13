# go-ve-sensor

# install tools
sudo npm install -g bower
go get -u github.com/jteeuwen/go-bindata/...

# setup dev environemnt
go generate
go build

# debugging
go get github.com/mailgun/godebug
godebug build -instrument=github.com/koestler/go-ve-sensor/vedirect