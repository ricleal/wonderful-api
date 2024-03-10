#!/bin/bash

# echo "GET http://localhost:8888/api/v1/wonderfuls?limit=100" | vegeta attack -rate 1000/s -duration=5s | tee results.bin | vegeta report

jq -ncM 'while(.<5000; .+1) | 
if (((.+1) % 100) != 0) then 
    {method: "GET", url: "http://localhost:8888/api/v1/wonderfuls?limit=\((.+1) % 100)" } 
else 
    empty
end' | \
    vegeta attack -rate=1000/s -lazy -format=json -duration=5s | \
    tee results.bin | \
    vegeta report

vegeta report -type=json results.bin > metrics.json
cat results.bin | vegeta plot > plot.html
cat results.bin | vegeta report -type="hist[0,1ms,2ms,3ms,4ms,5ms,6ms,7ms,8ms,9ms,10ms]" > hist.txt
cat hist.txt

