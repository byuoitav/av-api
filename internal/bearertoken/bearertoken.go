package bearertoken

import (
	"crypto/subtle"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Token struct {
	Token string `json:"token"`
}

func GetToken() (Token, error) {
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	response, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("elasticbeanstalk-us-west-2-194925301021"),
		Key:    aws.String("bearer-token.json"),
	})

	if err != nil {
		return Token{}, err
	}

	defer response.Body.Close()

	result := Token{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return Token{}, err
	}

	return result, nil
}

func CheckToken(tokenToCheck []byte) (bool, error) {
	token, err := GetToken()
	if err != nil {
		return false, err
	}

	tokenBytes := []byte(token.Token)

	result := subtle.ConstantTimeCompare(tokenBytes, tokenToCheck)
	if result == 1 {
		return true, nil
	}

	return false, nil
}
