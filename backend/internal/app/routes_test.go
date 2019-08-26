package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r-cbb/cbbpoll/internal/db/mocks"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

var testTeam = models.Team {
	FullName:   "University of Arizona",
	ShortName:  "Arizona",
	Nickname:   "Wildcats",
	Conference: "Pac-12",
}

var returnedTeam = models.Team{
	ID: 1,
	FullName:   "University of Arizona",
	ShortName:  "Arizona",
	Nickname:   "Wildcats",
	Conference: "Pac-12",
}

func addTeamMockDb() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", testTeam).Return(returnedTeam, nil).Once()
	return myMock
}

func addTeamDbError() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", testTeam).Return(models.Team{}, fmt.Errorf("some error")).Once()
	return myMock
}

func addTeamConcurrencyError() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", testTeam).Return(models.Team{}, errors.E(errors.KindConcurrencyProblem, fmt.Errorf("some error"))).Once()
	myMock.On("AddTeam", testTeam).Return(returnedTeam, nil).Once()
	return myMock
}

func TestAddTeam(t *testing.T) {
	testTeamJson, err := json.Marshal(returnedTeam)
	if err != nil {
		panic("Couldn't marshal testTeam")
	}
	testTeamStr := string(testTeamJson) + "\n"

	tests := []struct {
		name           string
		input          interface{}
		expectedStatus int
		expectedBody   string
		mockDb         mocks.DBClient
	}{
		{
			name:           "Successful add",
			input:          testTeam,
			expectedStatus: http.StatusOK,
			expectedBody:   testTeamStr,
			mockDb:         addTeamMockDb(),
		},
		{
			name:           "Bad input",
			input:          "{{{{foo%%",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Database error",
			input:          testTeam,
			expectedStatus: http.StatusInternalServerError,
			mockDb:         addTeamDbError(),
		},
		{
			name:           "Concurrency Retry",
			input:          testTeam,
			expectedStatus: http.StatusOK,
			expectedBody:   testTeamStr,
			mockDb:         addTeamConcurrencyError(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(test.input)
			if err != nil {
				t.Error(err)
				return
			}

			r := httptest.NewRequest(http.MethodPost, "/teams", &buf)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("AddTeam returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
				return
			}

			bodyBytes, err := ioutil.ReadAll(w.Body)
			bs := string(bodyBytes)

			if bs != test.expectedBody {
				t.Errorf("Response body differs from expected. Expected: %v, Actual: %v", test.expectedBody, bs)
			}

			test.mockDb.AssertExpectations(t)

			return
		})

	}

}
