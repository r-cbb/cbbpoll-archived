package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r-cbb/cbbpoll/backend/internal/db/mocks"
	"github.com/r-cbb/cbbpoll/backend/internal/errors"
	"github.com/r-cbb/cbbpoll/backend/pkg"
)

var testTeam = pkg.Team{
	FullName:   "University of Arizona",
	ShortName:  "Arizona",
	Nickname:   "Wildcats",
	Conference: "Pac-12",
}

func addTeamMockDb() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", testTeam).Return(int64(1), nil).Once()
	return myMock
}

func addTeamDbError() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", testTeam).Return(int64(0), fmt.Errorf("some error")).Once()
	return myMock
}

func addTeamConcurrencyError() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", testTeam).Return(int64(0), errors.E(errors.KindConcurrencyProblem, fmt.Errorf("some error"))).Times(1)
	myMock.On("AddTeam", testTeam).Return(int64(1), nil).Once()
	return myMock
}

func TestAddTeam(t *testing.T) {
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
			expectedBody:   "1\n",
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
			expectedBody:   "1\n",
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

			r := httptest.NewRequest(http.MethodPost, "/team", &buf)
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
