package model

import (
	"database/sql"
	"time"

	"github.com/motki/motkid/eveapi"
)

type Alliance struct {
	AllianceID  int
	Name        string
	DateFounded time.Time
	Ticker      string
}

func (m *Manager) GetAlliance(allianceID int) (*Alliance, error) {
	c, err := m.getAllianceFromDB(allianceID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return m.getAllianceFromAPI(allianceID)
	}
	return c, nil
}

func (m *Manager) getAllianceFromDB(allianceID int) (*Alliance, error) {
	c, err := m.pool.Open()
	if err != nil {
		return nil, err
	}
	defer c.Close()
	r := c.QueryRow(
		`SELECT
			  c.alliance_id
			, c.name
			, c.founded_date
			, c.ticker
			FROM app.alliances c
			WHERE c.alliance_id = $1`, allianceID)
	char := &Alliance{}
	err = r.Scan(
		&char.AllianceID,
		&char.Name,
		&char.DateFounded,
		&char.Ticker,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return char, nil
}

func (m *Manager) getAllianceFromAPI(allianceID int) (*Alliance, error) {
	char, err := m.eveapi.GetAlliance(allianceID)
	if err != nil {
		return nil, err
	}
	return m.apiAllianceToDB(char)
}

func (m *Manager) apiAllianceToDB(alliance *eveapi.Alliance) (*Alliance, error) {
	db, err := m.pool.Open()
	if err != nil {
		return nil, err
	}
	c := &Alliance{
		AllianceID:  alliance.AllianceID,
		Name:        alliance.Name,
		DateFounded: alliance.DateFounded,
		Ticker:      alliance.Ticker,
	}
	_, err = db.Exec(
		"INSERT INTO app.alliances (alliance_id, name, founded_date, ticker) VALUES($1, $2, $3, $4)",
		c.AllianceID,
		c.Name,
		c.DateFounded,
		c.Ticker,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}