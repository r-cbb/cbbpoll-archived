package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	authMocks "github.com/r-cbb/cbbpoll/internal/auth/mocks"
	"github.com/r-cbb/cbbpoll/internal/db/mocks"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

var inputTeam = models.Team{
	FullName:   "University of Arizona",
	ShortName:  "Arizona",
	Nickname:   "Wildcats",
	Conference: "Pac-12",
}

var testArizona = models.Team{
	ID:         1,
	FullName:   "University of Arizona",
	ShortName:  "Arizona",
	Nickname:   "Wildcats",
	Conference: "Pac-12",
}

var testOhioState = models.Team{
	ID:         2,
	FullName:   "Ohio State University",
	ShortName:  "Ohio State",
	Nickname:   "Buckeyes",
	Conference: "Big-10",
}

var testAdmin = models.User{
	Nickname: "Concision",
	IsAdmin:  true,
}

var testUser = models.User{
	Nickname: "JohnDoe",
	IsAdmin:  false,
}

func addTeamMockDb() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", inputTeam).Return(testArizona, nil).Once()
	return myMock
}

func addTeamDbError() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", inputTeam).Return(models.Team{}, fmt.Errorf("some error")).Once()
	return myMock
}

func addTeamConcurrencyError() mocks.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddTeam", inputTeam).Return(models.Team{}, errors.E(errors.KindConcurrencyProblem, fmt.Errorf("some error"))).Once()
	myMock.On("AddTeam", inputTeam).Return(testArizona, nil).Once()
	return myMock
}

func TestAddTeam(t *testing.T) {
	testTeamJson, err := json.Marshal(testArizona)
	if err != nil {
		panic("Couldn't marshal inputTeam")
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
			input:          inputTeam,
			expectedStatus: http.StatusCreated,
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
			input:          inputTeam,
			expectedStatus: http.StatusInternalServerError,
			mockDb:         addTeamDbError(),
		},
		{
			name:           "Concurrency Retry",
			input:          inputTeam,
			expectedStatus: http.StatusCreated,
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

			r := httptest.NewRequest(http.MethodPost, "/v1/teams", &buf)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("AddTeam returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
				return
			}

			test.mockDb.AssertExpectations(t)

			return
		})
	}
}

func TestPing(t *testing.T) {
	srv := NewServer()

	r := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected StatusOK, got: %v %v", w.Result().StatusCode, w.Result().Status)
	}

	bodyBytes, _ := ioutil.ReadAll(w.Body)
	bs := string(bodyBytes)

	expected, _ := json.Marshal(models.VersionInfo{Version: srv.version()})
	expStr := string(expected) + "\n"

	if bs != expStr {
		t.Errorf("Response body differs.  Expected: %v, Got: %v", bs, expStr)
	}
}

func TestGetMe(t *testing.T) {
	getDb := func(nick string, user models.User, err error) mocks.DBClient {
		myMock := mocks.DBClient{}
		myMock.On("GetUser", nick).Return(user, err)
		return myMock
	}

	tests := []struct {
		name           string
		expectedStatus int
		mockDb         mocks.DBClient
		authClient     authMocks.AuthClient
	}{
		{
			name:           "Success",
			expectedStatus: http.StatusOK,
			mockDb:         getDb("Concision", models.User{Nickname: "Concision", IsAdmin: true}, nil),
			authClient:     getAuth(models.UserToken{Nickname: "Concision"}),
		},
		{
			name:           "Not logged in",
			expectedStatus: http.StatusUnauthorized,
			authClient:     getAuth(models.UserToken{}),
		},
		{
			name:           "Database error",
			expectedStatus: http.StatusInternalServerError,
			mockDb:         getDb("Concision", models.User{}, fmt.Errorf("Some error")),
			authClient:     getAuth(models.UserToken{Nickname: "Concision"}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb
			srv.AuthClient = &test.authClient

			r := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("users/me returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
				return
			}
		})
	}
}

func TestGetTeam(t *testing.T) {
	getDb := func(id int64, team models.Team, err error) mocks.DBClient {
		myMock := mocks.DBClient{}
		myMock.On("GetTeam", id).Return(team, err)
		return myMock
	}

	tests := []struct {
		name           string
		expectedStatus int
		mockDb         mocks.DBClient
	}{
		{
			name:           "Success",
			expectedStatus: http.StatusOK,
			mockDb:         getDb(int64(1), testArizona, nil),
		},
		{
			name:           "Not found",
			expectedStatus: http.StatusNotFound,
			mockDb:         getDb(int64(1), models.Team{}, errors.E(errors.KindNotFound)),
		},
		{
			name:           "DB Error",
			expectedStatus: http.StatusInternalServerError,
			mockDb:         getDb(int64(1), models.Team{}, errors.E()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb

			r := httptest.NewRequest(http.MethodGet, "/v1/teams/1", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("/teams/1 returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
			}

			if !testSuccess(w.Result().StatusCode) {
				return
			}

			var res models.Team
			err := json.NewDecoder(w.Body).Decode(&res)
			if err != nil {
				t.Errorf("Error decoding json response: %v", err.Error())
			}

			if res != testArizona {
				t.Errorf("Expected Team  %v, got %v", testArizona, res)
			}

		})
	}
}

func TestListTeams(t *testing.T) {
	getDb := func(teams []models.Team, err error) mocks.DBClient {
		myMock := mocks.DBClient{}
		myMock.On("GetTeams").Return(teams, err)
		return myMock
	}

	tests := []struct {
		name           string
		expectedStatus int
		mockDb         mocks.DBClient
		expectedTeams  []models.Team
	}{
		{
			name:           "No Teams",
			expectedStatus: http.StatusOK,
			mockDb:         getDb(nil, nil),
		},
		{
			name:           "One Team",
			expectedStatus: http.StatusOK,
			mockDb:         getDb([]models.Team{testArizona}, nil),
			expectedTeams:  []models.Team{testArizona},
		},
		{
			name:           "Two Teams",
			expectedStatus: http.StatusOK,
			mockDb:         getDb([]models.Team{testArizona, testOhioState}, nil),
			expectedTeams:  []models.Team{testArizona, testOhioState},
		},
		{
			name:           "Database Error",
			expectedStatus: http.StatusInternalServerError,
			mockDb:         getDb(nil, errors.E()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb

			r := httptest.NewRequest(http.MethodGet, "/v1/teams", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("/teams returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
			}

			if !testSuccess(w.Result().StatusCode) {
				return
			}

			var res []models.Team
			err := json.NewDecoder(w.Body).Decode(&res)
			if err != nil {
				t.Errorf("Error decoding json response: %v", err.Error())
			}

			if !reflect.DeepEqual(res, test.expectedTeams) {
				t.Errorf("Expected Teams %v, got %v", test.expectedTeams, res)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	getDb := func(nick string, user models.User, err error) mocks.DBClient {
		myMock := mocks.DBClient{}
		myMock.On("GetUser", nick).Return(user, err)
		return myMock
	}

	tests := []struct {
		name           string
		expectedStatus int
		mockDb         mocks.DBClient
		expectedUser   models.User
	}{
		{
			name:           "OK",
			expectedStatus: http.StatusOK,
			mockDb:         getDb(testUser.Nickname, testUser, nil),
			expectedUser:   testUser,
		},
		{
			name:           "Not found",
			expectedStatus: http.StatusNotFound,
			mockDb:         getDb(testUser.Nickname, models.User{}, errors.E(errors.KindNotFound)),
		},
		{
			name:           "Database Error",
			expectedStatus: http.StatusInternalServerError,
			mockDb:         getDb(testUser.Nickname, models.User{}, errors.E()),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s", testUser.Nickname), nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("/v1/users/%s returned %v, expected %v", testUser.Nickname, w.Result().StatusCode, test.expectedStatus)
			}

			if !testSuccess(w.Result().StatusCode) {
				return
			}

			var res models.User
			err := json.NewDecoder(w.Body).Decode(&res)
			if err != nil {
				t.Errorf("Error decoding json response: %v", err.Error())
			}

			if res != test.expectedUser {
				t.Errorf("Expected User %v, got %v", test.expectedUser, res)
			}
		})
	}
}

type mockRedditClient struct {
	token string
	name  string
	err   error
}

func (c mockRedditClient) UsernameFromToken(token string) (name string, err error) {
	if token != c.token {
		panic("tokens don't match")
	}
	return c.name, c.err
}

func newMockRedditClient(expToken string, name string, err error) mockRedditClient {
	return mockRedditClient{
		token: expToken,
		name:  name,
		err:   err,
	}
}

func TestNewSession(t *testing.T) {
	const redditToken = "some.reddit.token"
	const expectedToken = "some.token.value"

	getDb := func(nick string, user models.User, err error, err2 error) mocks.DBClient {
		myMock := mocks.DBClient{}
		myMock.On("GetUser", nick).Return(user, err)
		myMock.On("AddUser", mock.AnythingOfType("models.User")).Return(user, err2)
		return myMock
	}

	getAuth := func(token string, err error) authMocks.AuthClient {
		myMock := authMocks.AuthClient{}
		myMock.On("CreateJWT", testUser).Return(token, err)
		myMock.On("Verifier").Return(func(next http.Handler) http.Handler {
			return http.Handler(next)
		})
		return myMock
	}

	tests := []struct {
		name           string
		expectedStatus int
		mockDb         mocks.DBClient
		mockAuth       authMocks.AuthClient
		redditClient   mockRedditClient
		redditToken    string
	}{
		{
			name:           "OK",
			expectedStatus: http.StatusOK,
			mockDb:         getDb(testUser.Nickname, testUser, nil, nil),
			mockAuth:       getAuth(expectedToken, nil),
			redditClient:   newMockRedditClient(redditToken, testUser.Nickname, nil),
			redditToken:    redditToken,
		},
		{
			name:           "BadRequest",
			expectedStatus: http.StatusBadRequest,
			mockDb:         getDb(testUser.Nickname, testUser, nil, nil),
			mockAuth:       getAuth(expectedToken, nil),
			redditClient:   newMockRedditClient(redditToken, testUser.Nickname, nil),
			redditToken:    "Bearer Bearer Bearer Token",
		},
		{
			name:           "Unauthorized Reddit Token",
			expectedStatus: http.StatusUnauthorized,
			mockDb:         getDb(testUser.Nickname, testUser, nil, nil),
			mockAuth:       getAuth(expectedToken, nil),
			redditClient:   newMockRedditClient(redditToken, testUser.Nickname, errors.E(errors.KindAuthError)),
			redditToken:    redditToken,
		},
		{
			name:           "Reddit Unavailable",
			expectedStatus: http.StatusServiceUnavailable,
			mockDb:         getDb(testUser.Nickname, testUser, nil, nil),
			mockAuth:       getAuth(expectedToken, nil),
			redditClient:   newMockRedditClient(redditToken, testUser.Nickname, errors.E(errors.KindServiceUnavailable)),
			redditToken:    redditToken,
		},
		{
			name:           "New User",
			expectedStatus: http.StatusCreated,
			mockDb:         getDb(testUser.Nickname, testUser, errors.E(errors.KindNotFound), nil),
			mockAuth:       getAuth(expectedToken, nil),
			redditClient:   newMockRedditClient(redditToken, testUser.Nickname, nil),
			redditToken:    redditToken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := NewServer()
			srv.Db = &test.mockDb
			srv.RedditClient = test.redditClient
			srv.AuthClient = &test.mockAuth
			srv.AuthRoutes()

			r := httptest.NewRequest(http.MethodPost, "/v1/sessions", nil)
			r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.redditToken))
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, r)
			if w.Result().StatusCode != test.expectedStatus {
				t.Errorf("POST /v1/sessions returned %v, expected %v", w.Result().StatusCode, test.expectedStatus)
			}

			if !testSuccess(w.Result().StatusCode) {
				return
			}
		})
	}
}

// Helpers

func getAuth(token models.UserToken) authMocks.AuthClient {
	myMock := authMocks.AuthClient{}
	myMock.On("UserTokenFromCtx", mock.Anything).Return(token)
	return myMock
}

func testSuccess(status int) bool {
	return status >= 200 && status < 300
}
