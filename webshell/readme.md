# test login API
curl -X POST -H "Content-Type: application/json" -d "{\"cluster\":\"yh9\",\"username\":\"root\",\"password\":\"nscc@123\",\"host\":\"1.94.239.51\"}" "http://localhost:8080/api/v2/document/login/"
{"sessionKey":"tsh_a2e932b625c0d598db3800aa91b92016"}

# test get document list API
curl -X GET -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" "http://localhost:8080/api/v2/document/list/?cluster=yh9"
