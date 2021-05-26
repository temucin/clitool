package utils

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

func createSession() *session.Session {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("Error creating session", err)
	}
	return sess
}

//GetInstances will take a set of filters and execute the Describe Instances command and return the results
func GetInstances(filters []*ec2.Filter) (*ec2.DescribeInstancesOutput, error) {
	sess := createSession()
	ec2Svc := ec2.New(sess, &aws.Config{Region: aws.String("us-east-1")}) //TODO(Get region from AWS config)
	describeParams := &ec2.DescribeInstancesInput{Filters: filters}
	return ec2Svc.DescribeInstances(describeParams)
}

//AssumeRole executes the assume command using the specified profile, RoleArn, and KeyId
func AssumeRole(roleArn string, profile string, keyID string) (*sts.AssumeRoleOutput, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
	})
	if err != nil {
		fmt.Println("Error creating session!", err)
	}

	stsSvc := sts.New(sess)
	roleSessionName := createSessionName(keyID) //We use the AwsAccessKeyId to create the session name to leave an audit trail
	assumeInput := sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &roleSessionName,
	}

	return stsSvc.AssumeRole(&assumeInput)
}

func createSessionName(keyID string) string {
	r := rand.New(rand.NewSource(99))
	return keyID + strconv.Itoa(r.Int())
}
