package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	authMocks "github.com/r-cbb/cbbpoll/internal/auth/mocks"
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

			bodyBytes, _ := ioutil.ReadAll(w.Body)
			bs := string(bodyBytes)

			if bs != test.expectedBody {
				t.Errorf("Response body differs from expected. Expected: %v, Actual: %v", test.expectedBody, bs)
			}

			test.mockDb.AssertExpectations(t)

			return
		})
	}
}

func TestPing(t *testing.T) {
	srv := NewServer()

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected StatusOK, got: %v %v", w.Result().StatusCode, w.Result().Status)
	}

	bodyBytes, _ := ioutil.ReadAll(w.Body)
	bs := string(bodyBytes)

	expected, _ := json.Marshal(struct{Version string}{Version: srv.version()})
	expStr := string(expected) + "\n"

	if bs != expStr {
		t.Errorf("Response body differs.  Expected: %v, Got: %v", bs, expStr)
	}
}

func getTeamMock() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetTeam", int64(23)).Return(testTeam, nil).Once()
	return myMock
}

func getTeamNotFoundMock() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetTeam", int64(23)).Return(models.Team{}, errors.E(errors.KindNotFound)).Once()
	return myMock
}

func getTeamErrorMock() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetTeam", int64(23)).Return(models.Team{}, fmt.Errorf("some error")).Once()
	return myMock
}

func Test_GetTeam(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedTeam   models.Team
		mockDb         mocks.DBClient
	}{
		{
			name:           "Success",
			expectedStatus: http.StatusOK,
			expectedTeam:   testTeam,
			mockDb:         getTeamMock(),
		},
		{
			name:           "Not Found",
			expectedStatus: http.StatusNotFound,
			mockDb:         getTeamNotFoundMock(),
		},
		{
			name:           "Database error",
			expectedStatus: http.StatusInternalServerError,
			mockDb:         getTeamErrorMock(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb

			r := httptest.NewRequest(http.MethodGet, "/teams/23", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("GetTeam returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
				return
			}
		})
	}
}

func Test_GetMe(t *testing.T) {
	getAuth := func(token models.UserToken) authMocks.AuthClient {
		myMock := authMocks.AuthClient{}
		myMock.On("UserTokenFromCtx", mock.Anything).Return(token)
		return myMock
	}

	getDb := func(nick string, user models.User, err error) mocks.DBClient {
		myMock := mocks.DBClient{}
		myMock.On("GetUser", nick).Return(user, err)
		return myMock
	}

	tests := []struct {
		name string
		expectedStatus int
		mockDb mocks.DBClient
		authClient authMocks.AuthClient
	} {
		{
			name: "Success",
			expectedStatus: http.StatusOK,
			mockDb: getDb("Concision", models.User{Nickname: "Concision", IsAdmin:true}, nil),
			authClient: getAuth(models.UserToken{Nickname: "Concision"}),
		},
		{
			name: "Not logged in",
			expectedStatus: http.StatusUnauthorized,
			authClient: getAuth(models.UserToken{}),
		},
		{
			name: "Database error",
			expectedStatus: http.StatusInternalServerError,
			mockDb: getDb("Concision", models.User{}, fmt.Errorf("Some error")),
			authClient: getAuth(models.UserToken{Nickname: "Concision"}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb
			srv.AuthClient = &test.authClient

			r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("users/me returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
				return
			}
		})
	}
}
