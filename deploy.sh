#!/usr/bin/env bash

configure_aws_cli(){
	aws --version
	aws configure set default.region us-east-1
	aws configure set default.output json
}

push_ecr_image(){
	eval $(aws ecr get-login --region us-east-1 --no-include-email)
	docker push $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/dsw:$CIRCLE_SHA1
}

configure_aws_cli
push_ecr_image
