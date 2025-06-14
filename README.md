# authproxy

## build
CGO_ENABLED=0 go build -o authproxy -ldflags="-s -w" *.go

## install
systemctl stop authproxy.service
cp -r authproxy profiles.yaml config /opt/authproxy 
systemctl start authproxy.service

## watch logs
journalctl --unit authproxy.service --since "15 minutes ago"