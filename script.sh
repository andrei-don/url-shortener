#!/bin/bash
while true; do
  response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://0.0.0.0:8080/shorten" \
    -H "Content-Type: application/json" \
    -d '{"url": "http://example.com"}')

  if [ "$response" -eq 200 ] || [ "$response" -eq 201 ]; then
    echo "Request succeeded with status code $response!"
    break
  fi

  echo "Request failed with status code $response. Retrying..."
  sleep 3
done
