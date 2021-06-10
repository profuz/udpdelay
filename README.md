# Running (in three separate terminals)

    go run main.go -i 225.0.0.1:1234 -o 225.0.0.1:1235 -delay 3
    ffmpeg -re -f lavfi -i testsrc=d=999999:size=720x576:rate=25 -f "lavfi" -i "sine=frequency=620:beep_factor=4:duration=999999:sample_rate=48000" -vcodec mpeg2video -preset veryfast -pix_fmt yuv420p -strict -2 -y -f mpegts -r 25 "udp://225.0.0.1:1234?ttl=3&pkt_size=1316"
    ffplay udp://225.0.0.1:1235

# Running linters and code formatters

    find . -name "*.go" -not -path ".git/*" | xargs gofmt -s -w
    $GOPATH/bin/golint -min_confidence 0.21 ./...

# Building for production

    CGO_ENABLED=0 go build -ldflags "-s -w"
