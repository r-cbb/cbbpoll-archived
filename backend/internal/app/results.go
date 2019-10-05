package app

import (
	"sort"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type resultsSlice []models.Result

func (rs resultsSlice) Len() int {
	return len(rs)
}

func (rs resultsSlice) Less(i, j int) bool {
	if rs[i].Points == rs[j].Points {
		return rs[i].TeamName < rs[j].TeamName
	}

	return rs[i].Points > rs[j].Points
}

func (rs resultsSlice) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (ps PollService) calcPollResults(poll models.Poll) ([]models.Result, error) {
	const op errors.Op = "app.calcPollResults"

	ballots, err := ps.Db.GetBallotsByPoll(poll)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving ballots associated with poll")
	}

	official := make([]models.Ballot, 0, len(ballots))
	for _, b := range ballots {
		if b.IsOfficial {
			official = append(official, b)
		}
	}

	officialResults, err := ps.resultsFromBallots(official)
	if err != nil {
		return nil, errors.E(op, err, "error calculating results from official ballots")
	}

	results, err := ps.resultsFromBallots(ballots)
	if err != nil {
		return nil, errors.E(op, err, "error calculating results from all ballots")
	}

	err = ps.Db.SetResults(poll, officialResults, results)
	if err != nil {
		return nil, errors.E(op, err, "error updating poll after calculating results")
	}

	return []models.Result(results), nil
}

func (ps PollService) resultsFromBallots(bs []models.Ballot) ([]models.Result, error) {
	resMap := make(map[int64]models.Result)
	for _, b := range bs {
		for _, vote := range b.Votes {
			res := resMap[vote.TeamID]
			if vote.Rank == 1 {
				res.FirstPlaceVotes = res.FirstPlaceVotes + 1
			}
			res.Points = res.Points + 26 - vote.Rank
			resMap[vote.TeamID] = res
		}
	}

	// Fill in team names and slugs
	keys := make([]int64, len(resMap))

	i := 0
	for k := range resMap {
		keys[i] = k
		i++
	}

	teams, err := ps.Db.GetTeamsByID(keys)
	if err != nil {
		return nil, errors.E(err, "unable to retrieve teams specified by ballot")
	}

	for _, t := range teams {
		mapTeam := resMap[t.ID]
		mapTeam.TeamName = t.ShortName
		if mapTeam.TeamName == "" {
			mapTeam.TeamName = t.FullName
		}
		mapTeam.TeamSlug = t.Slug

		resMap[t.ID] = mapTeam
	}

	results := make(resultsSlice, len(resMap))
	i = 0
	for k, v := range resMap {
		v.TeamID = k
		results[i] = v
		i++
	}

	sort.Sort(results)

	// Assign ranks
	for i := range results {
		results[i].Rank = i + 1
		if results[i].Rank > 25 {
			results[i].Rank = 0
		}

		// If a team has the same number of points as the one before it, it
		// shares the same rank
		if i > 0 && results[i].Points == results[i-1].Points {
			results[i].Rank = results[i-1].Rank
		}
	}

	return []models.Result(results), nil
}
