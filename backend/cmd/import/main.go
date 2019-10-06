package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/r-cbb/cbbpoll/internal/models"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/r-cbb/cbbpoll/internal/app"
	"github.com/r-cbb/cbbpoll/internal/db/sqlite"
)

var layout = "2006-01-02 15:04:05"

func main() {
	var dbPath = flag.String("db-file", "data/cbbpoll.db", "location of sqlite database file")
	var teamPath = flag.String("teams-file", "dumps/teams.tsv", "location of tsv with team data")
	var userPath = flag.String("users-file", "dumps/users.tsv", "location of tsv with user data")
	var pollPath = flag.String("polls-file", "dumps/polls.tsv", "location of tsv with poll data")
	var ballotPath = flag.String("ballots-file", "dumps/ballots.tsv", "location of tsv with ballot data")
	var eventPath = flag.String("events-file", "dumps/voter_events.tsv", "location of voter events tsv")
	var votePath = flag.String("votes-file", "dumps/votes.tsv", "location of votes tsv")


	// Setup Database connection
	db, err := sqlite.NewClient(*dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	log.Println("Sqlite Client initialized")

	// Setup service layer
	a := app.NewPollService(db)

	teamFile, err := os.Open(*teamPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer teamFile.Close()

	userFile, err := os.Open(*userPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer userFile.Close()

	pollFile, err := os.Open(*pollPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pollFile.Close()

	ballotFile, err := os.Open(*ballotPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer ballotFile.Close()

	eventFile, err := os.Open(*eventPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer eventFile.Close()

	voteFile, err := os.Open(*votePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer voteFile.Close()

	token := models.UserToken{
		Nickname: "Concision",
		IsAdmin: true,
	}

	addTeams(teamFile, token, a)

	uMap := addUsers(userFile, token, a)

	pollMap := addPolls(pollFile, token, a)

	ballots := ballotsFromFile(ballotFile, uMap, pollMap)

	fillInVoterStatus(eventFile, ballots, uMap, a)

	ballots = fillInVotes(voteFile, ballots)

	addBallots(ballots, token, a)

	return
}

func addTeams(teamFile io.Reader, token models.UserToken, a *app.PollService) {
	r := csv.NewReader(teamFile)
	r.Comma = '\t'
	r.FieldsPerRecord = 7
	_, err := r.Read()
	if err == io.EOF {
		log.Println("Empty file")
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("error parsing ID field: %s", record[0])
		}

		team := models.Team{
			ID: int64(id),
			FullName: record[1],
			ShortName: record[2],
			Slug: record[6],
			Nickname: record[4],
			Conference: record[5],
		}

		t, err := a.AddTeam(token, team)
		if err != nil {
			log.Fatal("error adding team")
			log.Fatal(err.Error())
		}
		log.Printf("Added team: %v", t)
	}
}

func addUsers(userFile io.Reader, token models.UserToken, a *app.PollService) userMap {
	r := csv.NewReader(userFile)
	r.Comma = '\t'
	r.FieldsPerRecord = 12
	_, err := r.Read()
	if err == io.EOF {
		log.Println("Empty file")
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	m := userMap{}
	m.IDToNick = make(map[int]string)
	m.nickToID = make(map[string]int)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		userTeam, err := strconv.Atoi(record[9])
		if err != nil {
			userTeam = 0
		}

		user := models.User {
			Nickname: record[1],
			IsAdmin: record[3] == "a",
			IsVoter: false,
			PrimaryTeam: int64(userTeam),
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal("bad id for user")
			os.Exit(1)
		}

		m.nickToID[user.Nickname] = id
		m.IDToNick[id] = user.Nickname

		u, err := a.AddUser(token, user)
		if err != nil {
			log.Println("error adding user")
			log.Fatal(err.Error())
		}
		log.Printf("Added user: %v", u)
	}

	return m
}

func addPolls(pollFile io.Reader, token models.UserToken, a *app.PollService) map[int]models.Poll {
	r := csv.NewReader(pollFile)
	r.Comma = '\t'
	r.FieldsPerRecord = 6
	_, err := r.Read()
	if err == io.EOF {
		log.Fatal("Empty file")
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	m := make(map[int]models.Poll)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		season, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatal(err.Error())
		}

		week, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatal(err.Error())
		}

		weekName := ""
		if week == 0 {
			weekName = "Preseason"
		}

		openTime, err := time.Parse(layout, record[3])
		if err != nil {
			log.Fatal(err.Error())
		}

		closeTime, err := time.Parse(layout, record[4])
		if err != nil {
			log.Fatal(err.Error())
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err.Error())
		}

		poll := models.Poll{
			Season: season,
			Week: week,
			OpenTime: openTime,
			CloseTime: closeTime,
			WeekName: weekName,
			RedditURL: record[5],
		}

		p, err := a.AddPoll(token, poll)

		m[id] = p

		if err != nil {
			log.Println("error adding poll")
			log.Fatal(err.Error())
		}
		log.Printf("Added poll: %v", p)
	}

	return m
}

type userMap struct {
	nickToID map[string]int
	IDToNick map[int]string
}

func ballotsFromFile(reader io.Reader, m userMap, pm map[int]models.Poll) []models.Ballot {
	r := csv.NewReader(reader)
	r.Comma = '\t'
	r.FieldsPerRecord = 4
	_, err := r.Read()
	if err == io.EOF {
		log.Println("Empty file")
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	bs := make([]models.Ballot, 0)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err.Error())
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err.Error())
		}

		pid, err := strconv.Atoi(record[3])
		if err != nil {
			log.Fatal(err.Error())
		}

		p := pm[pid]

		uid, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatal(err.Error())
		}

		updatedTime, err := time.Parse(layout, record[1])
		if err != nil {
			log.Fatal(err.Error())
		}

		ballot := models.Ballot{
			ID: int64(id),
			PollSeason: p.Season,
			PollWeek: p.Week,
			UpdatedTime: updatedTime,
			User: m.IDToNick[uid],
			Votes: nil,
			IsOfficial: false,
		}

		bs = append(bs, ballot)
	}

	return bs
}

func fillInVoterStatus(eventFile io.Reader, bs []models.Ballot, um userMap, a *app.PollService) {
	type event struct {
		voter bool
		effective time.Time
	}

	events := make(map[string][]event)

	r := csv.NewReader(eventFile)
	r.Comma = '\t'
	r.FieldsPerRecord = 4
	_, err := r.Read()
	if err == io.EOF {
		log.Println("Empty file")
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err.Error())
		}

		uID, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatal(err.Error())
		}

		isVoter, err := strconv.ParseBool(record[3])
		if err != nil {
			log.Fatal(err.Error())
		}

		effective, err := time.Parse(layout, record[1])
		if err != nil {
			log.Fatal(err.Error())
		}

		events[um.IDToNick[uID]] = append(events[um.IDToNick[uID]], event{voter:isVoter, effective:effective})
	}

	for i := range bs {
		user := bs[i].User

		p, err := a.GetPoll(bs[i].PollSeason, bs[i].PollWeek)
		if err != nil {
			log.Fatal(err.Error())
		}

		userEvents := events[user]
		if len(userEvents) == 0 {
			bs[i].IsOfficial = false
			continue
		}

		currStatus := false
		for _, e := range userEvents {
			if e.effective.Before(p.CloseTime) {
				fmt.Println("voter status is:")
				fmt.Println(e.voter)
				currStatus = e.voter
			}
		}

		bs[i].IsOfficial = currStatus
	}
}

func fillInVotes(voteFile io.Reader, bs []models.Ballot) []models.Ballot {
	bMap := make(map[int64]models.Ballot)
	for i, b := range bs {
		bMap[b.ID] = bs[i]
	}

	s := bufio.NewScanner(voteFile)
	s.Scan()
	for s.Scan() {
		record := strings.Split(s.Text(), "\t")

		ballotID, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatal(err.Error())
		}

		teamID, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatal(err.Error())
		}

		rank, err := strconv.Atoi(record[3])
		if err != nil {
			log.Fatal(err.Error())
		}

		b := bMap[int64(ballotID)]
		var reason string
		if len(record) < 5 {
			reason = ""
		} else {
			reason = record[4]
		}
		b.Votes = append(b.Votes, models.Vote{TeamID: int64(teamID), Rank: rank, Reason: reason})
		bMap[int64(ballotID)] = b
	}

	res := make([]models.Ballot, 0, 500)
	for _, v := range bMap {
		res = append(res, v)
	}

	return res
}

func addBallots(bs []models.Ballot, token models.UserToken, a *app.PollService) {
	for _, b := range bs {
		if len(b.Votes) == 0 {
			fmt.Printf("empty ballot: %v", b.ID)
			continue
		}
		if len(b.Votes) > 25 {
			b.Votes = b.Votes[:25]
		}
		if len(b.Votes) == 50 {
			fmt.Printf("that fucky ballot with 50: %v\n", b.ID)
			continue
		}
		if len(b.Votes) == 100 {
			fmt.Printf("the extra fucky ballot with 100: %v\n", b.ID)
		}
		_, err := a.AddBallot(token, b)
		if err != nil {
			log.Println("Error adding ballot")
			log.Println(err.Error())
			log.Println(b)
			log.Println("---")
		}
	}
}