#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

sleep 2
echo ">>> start test"
curl "http://localhost:50500/api?key=Alice" &
curl "http://localhost:50500/api?key=Bob" &
curl "http://localhost:50500/api?key=Bob" &
curl "http://localhost:50500/api?key=Bob" &
curl "http://localhost:50500/api?key=Bob" &
curl "http://localhost:50500/api?key=Bob" &
curl "http://localhost:50500/api?key=Cindy" &
curl "http://localhost:50500/api?key=David" &
curl "http://localhost:50500/api?key=Eric" &
curl "http://localhost:50500/api?key=Eric" &
curl "http://localhost:50500/api?key=Eric" &
curl "http://localhost:50500/api?key=Frank" &


wait