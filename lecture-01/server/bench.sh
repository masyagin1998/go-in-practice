#!/usr/bin/env bash
#
# Benchmark for unified TCP servers.
#
# Usage: ./bench.sh <fib|sleep> [connections] [N] [port]
#   mode         — "fib" or "sleep"
#   connections  — number of parallel requests (default: 100000)
#   N            — Fibonacci number (fib mode only, default: 30)
#   port         — server port (default: 9001)
#
# Examples:
#   ./bench.sh fib 100000 30
#   ./bench.sh sleep 100000

MODE="${1:?Usage: $0 <fib|sleep> [connections] [N] [port]}"
CONNS="${2:-100000}"
N="${3:-30}"
PORT="${4:-9001}"

if [ "$MODE" = "fib" ]; then
    PAYLOAD="$N"
    echo "Benchmarking: $CONNS parallel requests, fib($N), port=$PORT"
else
    PAYLOAD="ping"
    echo "Benchmarking: $CONNS parallel requests, sleep 100ms, port=$PORT"
fi
echo "---"

time seq "$CONNS" | xargs -P"$CONNS" -I{} sh -c 'echo '"$PAYLOAD"' | nc -N localhost '"$PORT"' > /dev/null'

echo "---"
echo "Done: $CONNS requests completed."
