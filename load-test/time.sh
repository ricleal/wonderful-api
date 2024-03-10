#!/bin/bash

# The script is a simple bash script that uses  curl  to make requests to the API. 
# It starts by getting the last 100 records from the API and then uses the last record's ID to get the next 100 records. 
# It continues this process until there are no more records to fetch. 
#
# The script also measures the time it takes to complete the test. 
# Running the test 
# To run the test, you can use the following command: 
# $ ./tests.sh
# The script will output the time it took to complete the test. 

start_time=$(date +%s%3N)

total_length=100

last_id=$(curl -s 'localhost:8888/api/v1/wonderfuls?limit=100'  | jq -r '. | last | .id')

while : ; do
    contents=$(curl -s "localhost:8888/api/v1/wonderfuls?limit=100&starting_after=$last_id")
    length=$(jq -r '. | length' <<< "$contents")

    total_length=$((total_length + length))
    # echo "Fetched $length records of a total of $total_length records ($last_id)"
    if [ "$length" -eq 0 ]; then
        break
    fi
    last_id=$(jq -r '. | last | .id' <<< "$contents")
done


end_time=$(date +%s%3N)

echo "It took $((end_time - start_time)) ms to complete the test"
 
