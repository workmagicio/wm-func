
GOOS=linux GOARCH=amd64 go build .
scp -i ~xukai/.ssh/ali-us-va-default-key.pem /Users/xukai/GolandProjects/workmagic/wm-tools/cmd/amazon_vendor_report/amazon_vendor_report ecs-user@10.10.2.183:/home/ecs-user/amazon_vendor_report
