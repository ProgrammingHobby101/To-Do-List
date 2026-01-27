@REM DEL bootstrap
@REM DEL go_lambda_test.zip
@REM sh build_exe.sh
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0 
go build -tags lambda.norpc -o bootstrap main.go
powershell -Command "Compress-Archive bootstrap -f go_lambda_to-do-list.zip"
aws lambda update-function-code --function-name to-do-list-api-login --zip-file fileb://go_lambda_to-do-list.zip --region us-east-2
@REM Compress-Archive bootstrap go_lambda_test3.zip
@REM git archive --format=zip --output=go_lambda_test.zip HEAD bootstrap