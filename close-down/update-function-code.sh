rm -rf close-down handler.zip
GOARCH=amd64 GOOS=linux go build
zip handler.zip ./close-down
aws lambda update-function-code \
  --region ap-northeast-2 \
  --function-name $1 \
  --zip-file fileb://./handler.zip \

rm -rf close-down handler.zip
