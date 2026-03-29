package main_goals

import (
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestMainGoalTeamManagement(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newMainGoalsRouter(conn)
	session := seedMainGoalAndLogin(t, env.DB, router)
	accountRepo := authpersistence.NewAccountRepository(conn)
	outsider := createScenarioAccountAndLogin(t, accountRepo, router, "main-goal-team-outsider", "main-goal-team-outsider-co", "password")

	createdTeam := createTeamForScenario(t, router, session.Token, "Main Goal Alpha FC", 2012, "Jalan Stadion Utama 10", "Bandung")
	if createdTeam.CompanyID != session.CompanyID {
		t.Fatalf("expected company_id %d, got %d", session.CompanyID, createdTeam.CompanyID)
	}
	if createdTeam.Name != "Main Goal Alpha FC" || createdTeam.FoundedYear != 2012 || createdTeam.HomebaseAddress != "Jalan Stadion Utama 10" || createdTeam.CityOfHomebaseAddress != "Bandung" {
		t.Fatalf("unexpected created team payload: %#v", createdTeam)
	}

	duplicateCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", session.Token, map[string]any{
		"name":                     "Main Goal Alpha FC",
		"founded_year":             2014,
		"homebase_address":         "Jalan Stadion Utama 11",
		"city_of_homebase_address": "Bandung",
	})
	if duplicateCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, duplicateCreate.Code, duplicateCreate.Body.String())
	}

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), session.Token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	gotTeam := decodeJSONResponse[teamPayload](t, getResponse)
	if gotTeam != createdTeam {
		t.Fatalf("expected fetched team %#v, got %#v", createdTeam, gotTeam)
	}

	listResponse := sendRequest(t, router, http.MethodGet, "/api/v1/teams", session.Token, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}
	teamsList := decodeJSONResponse[itemsResponse[teamPayload]](t, listResponse)
	foundTeam := false
	for _, item := range teamsList.Items {
		if item.ID == createdTeam.ID {
			foundTeam = true
			break
		}
	}
	if !foundTeam {
		t.Fatalf("expected team %d in list, got %#v", createdTeam.ID, teamsList.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), session.Token, map[string]any{
		"name":                     "Main Goal Alpha FC Updated",
		"founded_year":             2012,
		"homebase_address":         "Jalan Stadion Utama 10",
		"city_of_homebase_address": "Jakarta",
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}
	updatedTeam := decodeJSONResponse[teamPayload](t, updateResponse)
	if updatedTeam.Name != "Main Goal Alpha FC Updated" || updatedTeam.CityOfHomebaseAddress != "Jakarta" {
		t.Fatalf("unexpected updated team payload: %#v", updatedTeam)
	}

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), outsider.Token, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), outsider.Token, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), session.Token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), session.Token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected deleted get status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	assertSoftDeletedAt(t, env.DB, "teams", createdTeam.ID)
}

func TestMainGoalPlayerManagement(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newMainGoalsRouter(conn)
	session := seedMainGoalAndLogin(t, env.DB, router)
	accountRepo := authpersistence.NewAccountRepository(conn)
	outsider := createScenarioAccountAndLogin(t, accountRepo, router, "main-goal-player-outsider", "main-goal-player-outsider-co", "password")

	team := createTeamForScenario(t, router, session.Token, "Main Goal Players FC", 2011, "Jalan Pemain 1", "Bandung")
	createdPlayer := createPlayerForScenario(t, router, session.Token, team.ID, "Main Goal Striker", 182.4, 77.1, "striker", 9)
	if createdPlayer.TeamID != team.ID || createdPlayer.Name != "Main Goal Striker" || createdPlayer.Position != "striker" || createdPlayer.PlayerNumber != 9 {
		t.Fatalf("unexpected created player payload: %#v", createdPlayer)
	}

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players/%d", team.ID, createdPlayer.ID), session.Token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	gotPlayer := decodeJSONResponse[playerPayload](t, getResponse)
	if gotPlayer != createdPlayer {
		t.Fatalf("expected fetched player %#v, got %#v", createdPlayer, gotPlayer)
	}

	listResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players", team.ID), session.Token, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}
	playersList := decodeJSONResponse[itemsResponse[playerPayload]](t, listResponse)
	if len(playersList.Items) != 1 || playersList.Items[0].ID != createdPlayer.ID {
		t.Fatalf("expected one player %d in list, got %#v", createdPlayer.ID, playersList.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/teams/%d/players/%d", team.ID, createdPlayer.ID), session.Token, map[string]any{
		"name":          "Main Goal Striker",
		"height":        182.4,
		"weight":        78.5,
		"position":      "midfielder",
		"player_number": 9,
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}
	updatedPlayer := decodeJSONResponse[playerPayload](t, updateResponse)
	if updatedPlayer.Position != "midfielder" || updatedPlayer.Weight != 78.5 {
		t.Fatalf("unexpected updated player payload: %#v", updatedPlayer)
	}

	secondPlayer := createPlayerForScenario(t, router, session.Token, team.ID, "Main Goal Defender", 185.0, 80.0, "defender", 5)
	if secondPlayer.PlayerNumber != 5 {
		t.Fatalf("expected second player number 5, got %d", secondPlayer.PlayerNumber)
	}

	listAfterSecondCreate := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players", team.ID), session.Token, nil)
	if listAfterSecondCreate.Code != http.StatusOK {
		t.Fatalf("expected second list status %d, got %d with body %s", http.StatusOK, listAfterSecondCreate.Code, listAfterSecondCreate.Body.String())
	}
	playersAfterSecondCreate := decodeJSONResponse[itemsResponse[playerPayload]](t, listAfterSecondCreate)
	if len(playersAfterSecondCreate.Items) != 2 {
		t.Fatalf("expected two active players in list, got %#v", playersAfterSecondCreate.Items)
	}

	invalidPosition := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", team.ID), session.Token, map[string]any{
		"name":          "Invalid Position",
		"height":        180.0,
		"weight":        72.0,
		"position":      "wingback",
		"player_number": 12,
	})
	if invalidPosition.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid position status %d, got %d with body %s", http.StatusBadRequest, invalidPosition.Code, invalidPosition.Body.String())
	}

	invalidNumber := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", team.ID), session.Token, map[string]any{
		"name":          "Invalid Number",
		"height":        181.0,
		"weight":        74.0,
		"position":      "defender",
		"player_number": 0,
	})
	if invalidNumber.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid number status %d, got %d with body %s", http.StatusBadRequest, invalidNumber.Code, invalidNumber.Body.String())
	}

	duplicateNumber := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", team.ID), session.Token, map[string]any{
		"name":          "Duplicate Number",
		"height":        179.0,
		"weight":        73.0,
		"position":      "goalkeeper",
		"player_number": 9,
	})
	if duplicateNumber.Code != http.StatusConflict {
		t.Fatalf("expected duplicate number status %d, got %d with body %s", http.StatusConflict, duplicateNumber.Code, duplicateNumber.Body.String())
	}

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players/%d", team.ID, createdPlayer.ID), outsider.Token, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d/players/%d", team.ID, createdPlayer.ID), outsider.Token, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d/players/%d", team.ID, createdPlayer.ID), session.Token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	assertSoftDeletedAt(t, env.DB, "players", createdPlayer.ID)

	reusedNumberPlayer := createPlayerForScenario(t, router, session.Token, team.ID, "Reused Number Player", 180.0, 75.0, "goalkeeper", 9)
	if reusedNumberPlayer.PlayerNumber != 9 {
		t.Fatalf("expected reused number 9, got %d", reusedNumberPlayer.PlayerNumber)
	}
}

func TestMainGoalScheduleManagement(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newMainGoalsRouter(conn)
	session := seedMainGoalAndLogin(t, env.DB, router)
	accountRepo := authpersistence.NewAccountRepository(conn)
	outsider := createScenarioAccountAndLogin(t, accountRepo, router, "main-goal-schedule-outsider", "main-goal-schedule-outsider-co", "password")

	homeTeam := createTeamForScenario(t, router, session.Token, "Home FC", 2010, "Jalan Home 1", "Jakarta")
	guestTeam := createTeamForScenario(t, router, session.Token, "Guest FC", 2011, "Jalan Guest 2", "Bandung")
	createdSchedule := createScheduleForScenario(t, router, session.Token, "2026-09-10", "15:30:00", homeTeam.ID, guestTeam.ID)
	if createdSchedule.CompanyID != session.CompanyID || createdSchedule.MatchDate != "2026-09-10" || createdSchedule.MatchTime != "15:30:00" || createdSchedule.HomeTeamID != homeTeam.ID || createdSchedule.GuestTeamID != guestTeam.ID {
		t.Fatalf("unexpected created schedule payload: %#v", createdSchedule)
	}

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), session.Token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	gotSchedule := decodeJSONResponse[schedulePayload](t, getResponse)
	if gotSchedule != createdSchedule {
		t.Fatalf("expected fetched schedule %#v, got %#v", createdSchedule, gotSchedule)
	}

	listResponse := sendRequest(t, router, http.MethodGet, "/api/v1/schedules", session.Token, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}
	schedulesList := decodeJSONResponse[itemsResponse[schedulePayload]](t, listResponse)
	if len(schedulesList.Items) != 1 || schedulesList.Items[0].ID != createdSchedule.ID {
		t.Fatalf("expected one schedule %d in list, got %#v", createdSchedule.ID, schedulesList.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), session.Token, map[string]any{
		"match_date":    "2026-09-12",
		"match_time":    "18:45:00",
		"home_team_id":  homeTeam.ID,
		"guest_team_id": guestTeam.ID,
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}
	updatedSchedule := decodeJSONResponse[schedulePayload](t, updateResponse)
	if updatedSchedule.MatchDate != "2026-09-12" || updatedSchedule.MatchTime != "18:45:00" {
		t.Fatalf("unexpected updated schedule payload: %#v", updatedSchedule)
	}

	secondSchedule := createScheduleForScenario(t, router, session.Token, "2026-09-19", "18:45:00", homeTeam.ID, guestTeam.ID)
	if secondSchedule.HomeTeamID != homeTeam.ID || secondSchedule.GuestTeamID != guestTeam.ID {
		t.Fatalf("unexpected second schedule payload: %#v", secondSchedule)
	}

	swappedSchedule := createScheduleForScenario(t, router, session.Token, "2026-09-19", "20:15:00", guestTeam.ID, homeTeam.ID)
	if swappedSchedule.HomeTeamID != guestTeam.ID || swappedSchedule.GuestTeamID != homeTeam.ID {
		t.Fatalf("unexpected swapped schedule payload: %#v", swappedSchedule)
	}

	duplicateCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", session.Token, map[string]any{
		"match_date":    updatedSchedule.MatchDate,
		"match_time":    updatedSchedule.MatchTime,
		"home_team_id":  homeTeam.ID,
		"guest_team_id": guestTeam.ID,
	})
	if duplicateCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, duplicateCreate.Code, duplicateCreate.Body.String())
	}

	sameTeamCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", session.Token, map[string]any{
		"match_date":    "2026-09-21",
		"match_time":    "19:00:00",
		"home_team_id":  homeTeam.ID,
		"guest_team_id": homeTeam.ID,
	})
	if sameTeamCreate.Code != http.StatusBadRequest {
		t.Fatalf("expected same-team create status %d, got %d with body %s", http.StatusBadRequest, sameTeamCreate.Code, sameTeamCreate.Body.String())
	}

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), outsider.Token, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), outsider.Token, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), session.Token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), session.Token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected deleted get status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	assertSoftDeletedAt(t, env.DB, "schedules", createdSchedule.ID)
}

func TestMainGoalReportManagement(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newMainGoalsRouter(conn)
	session := seedMainGoalAndLogin(t, env.DB, router)
	accountRepo := authpersistence.NewAccountRepository(conn)
	outsider := createScenarioAccountAndLogin(t, accountRepo, router, "main-goal-report-outsider", "main-goal-report-outsider-co", "password")

	homeTeam := createTeamForScenario(t, router, session.Token, "Report Home FC", 2010, "Jalan Report Home", "Bogor")
	guestTeam := createTeamForScenario(t, router, session.Token, "Report Guest FC", 2011, "Jalan Report Guest", "Depok")
	topScorer := createPlayerForScenario(t, router, session.Token, homeTeam.ID, "Home Top Scorer", 181.0, 76.0, "striker", 10)
	scheduleOne := createScheduleForScenario(t, router, session.Token, "2026-10-01", "15:00:00", homeTeam.ID, guestTeam.ID)
	createdReport := createReportForScenario(t, router, session.Token, scheduleOne.ID, 3, 1, &topScorer.ID)
	if createdReport.MatchScheduleID != scheduleOne.ID || createdReport.HomeTeamID != homeTeam.ID || createdReport.GuestTeamID != guestTeam.ID {
		t.Fatalf("unexpected created report payload: %#v", createdReport)
	}
	if createdReport.StatusMatch != "home_team_win" {
		t.Fatalf("expected status_match home_team_win, got %q", createdReport.StatusMatch)
	}
	if createdReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 1 || createdReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 0 {
		t.Fatalf("unexpected initial accumulate counters: %#v", createdReport)
	}

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), session.Token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	gotReport := decodeJSONResponse[reportPayload](t, getResponse)
	if gotReport.ID != createdReport.ID ||
		gotReport.CompanyID != createdReport.CompanyID ||
		gotReport.MatchScheduleID != createdReport.MatchScheduleID ||
		gotReport.HomeTeamID != createdReport.HomeTeamID ||
		gotReport.GuestTeamID != createdReport.GuestTeamID ||
		gotReport.FinalScoreHome != createdReport.FinalScoreHome ||
		gotReport.FinalScoreGuest != createdReport.FinalScoreGuest ||
		gotReport.StatusMatch != createdReport.StatusMatch ||
		gotReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != createdReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule ||
		gotReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != createdReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule {
		t.Fatalf("expected fetched report %#v, got %#v", createdReport, gotReport)
	}
	if gotReport.MostScoringGoalPlayerID == nil || createdReport.MostScoringGoalPlayerID == nil || *gotReport.MostScoringGoalPlayerID != *createdReport.MostScoringGoalPlayerID {
		t.Fatalf("expected fetched report top scorer %v, got %v", createdReport.MostScoringGoalPlayerID, gotReport.MostScoringGoalPlayerID)
	}

	listResponse := sendRequest(t, router, http.MethodGet, "/api/v1/reports", session.Token, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}
	reportsList := decodeJSONResponse[itemsResponse[reportPayload]](t, listResponse)
	if len(reportsList.Items) != 1 || reportsList.Items[0].ID != createdReport.ID {
		t.Fatalf("expected one report %d in list, got %#v", createdReport.ID, reportsList.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), session.Token, map[string]any{
		"match_schedule_id":           scheduleOne.ID,
		"final_score_home":            1,
		"final_score_guest":           1,
		"most_scoring_goal_player_id": nil,
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}
	updatedReport := decodeJSONResponse[reportPayload](t, updateResponse)
	if updatedReport.StatusMatch != "draw" || updatedReport.MostScoringGoalPlayerID != nil {
		t.Fatalf("unexpected updated report payload: %#v", updatedReport)
	}
	if updatedReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 0 || updatedReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 0 {
		t.Fatalf("expected draw accumulate counters to reset to zero, got %#v", updatedReport)
	}

	scheduleTwo := createScheduleForScenario(t, router, session.Token, "2026-10-08", "17:30:00", homeTeam.ID, guestTeam.ID)
	guestWinReport := createReportForScenario(t, router, session.Token, scheduleTwo.ID, 0, 2, nil)
	if guestWinReport.StatusMatch != "guest_team_win" {
		t.Fatalf("expected guest_team_win status, got %q", guestWinReport.StatusMatch)
	}
	if guestWinReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 1 {
		t.Fatalf("expected guest accumulated wins 1, got %d", guestWinReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule)
	}

	duplicateCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", session.Token, map[string]any{
		"match_schedule_id":           scheduleTwo.ID,
		"final_score_home":            1,
		"final_score_guest":           0,
		"most_scoring_goal_player_id": nil,
	})
	if duplicateCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, duplicateCreate.Code, duplicateCreate.Body.String())
	}

	thirdTeam := createTeamForScenario(t, router, session.Token, "Report Third FC", 2014, "Jalan Report Third", "Bekasi")
	outsidePlayer := createPlayerForScenario(t, router, session.Token, thirdTeam.ID, "Outside Player", 179.0, 72.0, "striker", 7)
	scheduleThree := createScheduleForScenario(t, router, session.Token, "2026-10-15", "19:00:00", homeTeam.ID, guestTeam.ID)
	outsideTopScorer := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", session.Token, map[string]any{
		"match_schedule_id":           scheduleThree.ID,
		"final_score_home":            2,
		"final_score_guest":           0,
		"most_scoring_goal_player_id": outsidePlayer.ID,
	})
	if outsideTopScorer.Code != http.StatusNotFound {
		t.Fatalf("expected outside top scorer status %d, got %d with body %s", http.StatusNotFound, outsideTopScorer.Code, outsideTopScorer.Body.String())
	}

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), outsider.Token, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), outsider.Token, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), session.Token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), session.Token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected deleted get status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	assertSoftDeletedAt(t, env.DB, "reports", createdReport.ID)
}

func TestMainGoalFullScenario(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newMainGoalsRouter(conn)
	session := seedMainGoalAndLogin(t, env.DB, router)

	teamA := createTeamForScenario(t, router, session.Token, "Team A Home", 2010, "Jalan Team A", "Jakarta")
	teamB := createTeamForScenario(t, router, session.Token, "Team B Guest", 2011, "Jalan Team B", "Bandung")
	playerOne := createPlayerForScenario(t, router, session.Token, teamA.ID, "Player One", 183.0, 77.0, "striker", 9)
	_ = createPlayerForScenario(t, router, session.Token, teamB.ID, "Player Two", 189.0, 82.0, "goalkeeper", 1)
	scheduleOne := createScheduleForScenario(t, router, session.Token, "2026-11-01", "10:00:00", teamA.ID, teamB.ID)
	scheduleTwo := createScheduleForScenario(t, router, session.Token, "2026-11-08", "10:00:00", teamA.ID, teamB.ID)

	reportOne := createReportForScenario(t, router, session.Token, scheduleOne.ID, 2, 0, &playerOne.ID)
	if reportOne.StatusMatch != "home_team_win" || reportOne.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 1 || reportOne.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 0 {
		t.Fatalf("unexpected first report payload: %#v", reportOne)
	}

	reportTwo := createReportForScenario(t, router, session.Token, scheduleTwo.ID, 1, 1, nil)
	if reportTwo.StatusMatch != "draw" || reportTwo.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 1 || reportTwo.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 0 {
		t.Fatalf("unexpected second report payload: %#v", reportTwo)
	}

	updateReportOne := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/reports/%d", reportOne.ID), session.Token, map[string]any{
		"match_schedule_id":           scheduleOne.ID,
		"final_score_home":            0,
		"final_score_guest":           1,
		"most_scoring_goal_player_id": nil,
	})
	if updateReportOne.Code != http.StatusOK {
		t.Fatalf("expected report update status %d, got %d with body %s", http.StatusOK, updateReportOne.Code, updateReportOne.Body.String())
	}
	updatedReportOne := decodeJSONResponse[reportPayload](t, updateReportOne)
	if updatedReportOne.StatusMatch != "guest_team_win" || updatedReportOne.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 0 || updatedReportOne.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 1 {
		t.Fatalf("unexpected updated first report payload: %#v", updatedReportOne)
	}

	getReportTwo := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", reportTwo.ID), session.Token, nil)
	if getReportTwo.Code != http.StatusOK {
		t.Fatalf("expected second report get status %d, got %d with body %s", http.StatusOK, getReportTwo.Code, getReportTwo.Body.String())
	}
	refreshedReportTwo := decodeJSONResponse[reportPayload](t, getReportTwo)
	if refreshedReportTwo.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 0 || refreshedReportTwo.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 1 {
		t.Fatalf("expected recomputed second report counters after first report update, got %#v", refreshedReportTwo)
	}

	deleteTeam := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d", teamA.ID), session.Token, nil)
	if deleteTeam.Code != http.StatusNoContent {
		t.Fatalf("expected team delete status %d, got %d with body %s", http.StatusNoContent, deleteTeam.Code, deleteTeam.Body.String())
	}

	getDeletedTeam := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", teamA.ID), session.Token, nil)
	if getDeletedTeam.Code != http.StatusNotFound {
		t.Fatalf("expected deleted team get status %d, got %d with body %s", http.StatusNotFound, getDeletedTeam.Code, getDeletedTeam.Body.String())
	}

	assertSoftDeletedAt(t, env.DB, "teams", teamA.ID)
	assertSoftDeletedAt(t, env.DB, "players", playerOne.ID)
	assertSoftDeletedAt(t, env.DB, "schedules", scheduleOne.ID)
	assertSoftDeletedAt(t, env.DB, "schedules", scheduleTwo.ID)
	assertSoftDeletedAt(t, env.DB, "reports", reportOne.ID)
	assertSoftDeletedAt(t, env.DB, "reports", reportTwo.ID)
}
