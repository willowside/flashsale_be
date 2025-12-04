# 用相同 user 在短時間內打 10 次
for i in {1..10}; do
  curl -s -X POST http://localhost:8080/flashsale/precheck \
    -H "Content-Type: application/json" \
    -H "X-User-ID: U1" \
    -d '{"product":"1001"}' | jq .
  echo
done
