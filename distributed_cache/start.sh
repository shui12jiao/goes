#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

sleep 2
echo ">>> start test"
curl "http://localhost:50500/api?key=Alice" -w '\n' &
curl "http://localhost:50500/api?key=Bob" -w '\n' &
curl "http://localhost:50500/api?key=Bob" -w '\n' &
curl "http://localhost:50500/api?key=Bob" -w '\n' &
curl "http://localhost:50500/api?key=Bob" -w '\n' &
curl "http://localhost:50500/api?key=Bob" -w '\n' &
curl "http://localhost:50500/api?key=Cindy" -w '\n' &
curl "http://localhost:50500/api?key=David" -w '\n' &
curl "http://localhost:50500/api?key=Eric" -w '\n' &
curl "http://localhost:50500/api?key=Eric" -w '\n' &
curl "http://localhost:50500/api?key=Eric" -w '\n' &
curl "http://localhost:50500/api?key=Frank" -w '\n' &


wait