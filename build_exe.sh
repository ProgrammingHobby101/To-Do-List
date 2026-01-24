#STEP #1: sh build_exe.sh
git add .
#git rm -f bootstrap
#git rm -f "go_lambda_test.zip"
GOOS=linux CGO_ENABLED=0 GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go
Compress-Archive bootstrap go_lambda_test3.zip
#git archive --format=zip --output="go_lambda_test.zip" HEAD bootstrap
# git mv -f bootstrap go_lambda_test.zip
# tar -cvf go_lambda_test.zip bootstrap
# cp -r bootstrap "/go_lambda_test.zip"
# cp -r bootstrap test