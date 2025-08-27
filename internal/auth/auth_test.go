package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

type makeInput struct {
	userID      uuid.UUID
	tokenSecret string
	expiresIn   time.Duration
}

type makeOutput struct {
	tokenString string
	errorOut    error
}

type makeCase struct {
	testInput  makeInput
	testOutput makeOutput
}

type validateInput struct {
	tokenString string
	tokenSecret string
}

type validateOutput struct {
	userID   uuid.UUID
	errorOut error
}

type validateCase struct {
	testInput  validateInput
	testOutput validateOutput
}

func TestMakeJWT(t *testing.T) {
	happyCase := makeCase{
		testInput: makeInput{
			userID:      uuid.New(),
			tokenSecret: "omgsecret",
			expiresIn:   time.Duration(time.Hour),
		},
		testOutput: makeOutput{},
	}

	happyCase.testOutput.tokenString, happyCase.testOutput.errorOut = MakeJWT(
		happyCase.testInput.userID,
		happyCase.testInput.tokenSecret,
		happyCase.testInput.expiresIn,
	)

	if happyCase.testOutput.errorOut != nil {
		t.Errorf("Got %v, expected nil", happyCase.testOutput.errorOut)
	}
}

func TestValidateJWT(t *testing.T) {
	happyCase := validateCase{
		testInput: validateInput{
			tokenSecret: "omgsecret",
		},
		testOutput: validateOutput{
			userID:   uuid.New(),
			errorOut: nil,
		},
	}
	happyCase.testInput.tokenString, _ = MakeJWT(
		happyCase.testOutput.userID,
		happyCase.testInput.tokenSecret,
		time.Duration(time.Hour),
	)

	testID, err := ValidateJWT(
		happyCase.testInput.tokenString,
		happyCase.testInput.tokenSecret,
	)
	if err != happyCase.testOutput.errorOut {
		t.Errorf("Error generated: Got %v, expected %v", err, happyCase.testOutput.errorOut)
	}

	if testID != happyCase.testOutput.userID {
		t.Errorf("Didn't get correct ID: Got %v, expected %v", testID, happyCase.testOutput.userID)
	}
}

func TestGetBearerToken(t *testing.T) {
	input := http.Header{}
	input = make(http.Header)

	input.Add("Authorization", "Bearer wobbles   ")

	output, err := GetBearerToken(input)
	if err != nil {
		t.Errorf("Error generated: Got %v, expected nil", err)
	}

	if output != "wobbles" {
		t.Errorf("Didn't get correct string: Got %s, expected wobbles", output)
	}
}
