package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"time"
)

func getEcrCredentials(awsAccount, awsRegion, awsAccessKeyID, awsSecretAccessKey string) (string, error) {
	if ExpiresAt.Add(-2*time.Hour).After(time.Now()){
		return configDockerconfigjson, nil
	}
	credentials := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")
	awsConfig := aws.NewConfig().WithRegion(awsRegion).WithCredentials(credentials)
	svc := ecr.New(session.New(), awsConfig)
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecr.ErrCodeServerException:
				fmt.Println(ecr.ErrCodeServerException, aerr.Error())
			case ecr.ErrCodeInvalidParameterException:
				fmt.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return "", err
	}

	fmt.Printf("%+v\n", result)
	base64EncodedToken := *result.AuthorizationData[0].AuthorizationToken
	dbuf := make([]byte, len(base64EncodedToken))
	count, err := base64.StdEncoding.Decode(dbuf, []byte(base64EncodedToken))
	if err != nil {
		return "", err
	}
	ExpiresAt = *result.AuthorizationData[0].ExpiresAt
	password := string(dbuf[4:count])
	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", awsAccount, awsRegion)
	dockerConfigJSON := map[string]interface{}{
		"auths": map[string]interface{}{
			registry:map[string]interface{}{
				"username":"AWS",
				"password":password,
			},
		},
	}


	dockerConfigBytes, err := json.Marshal(dockerConfigJSON)
	if err != nil {
		return "", err
	}
	configDockerconfigjson = string(dockerConfigBytes)
	dockerConfigBase64Bytes := make([]byte, len(dockerConfigBytes)*2)
	base64.StdEncoding.Encode(dockerConfigBase64Bytes, dockerConfigBytes)
	return string(dockerConfigBase64Bytes), nil
}
