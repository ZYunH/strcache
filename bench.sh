#!/bin/bash

timestamp=$(date +%T)
echo -n > "bench_${timestamp}.txt"

go test -bench=. -benchmem -count=10 > "bench_${timestamp}.txt"