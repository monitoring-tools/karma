// Package v04 package implements support for interacting with Alertmanager 0.4
// Collected data will be mapped to unsee internal schema defined the
// unsee/models package
// This file defines Alertmanager silences mapping
package v04

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/blang/semver"
	"github.com/cloudflare/unsee/internal/mapper"
	"github.com/cloudflare/unsee/internal/models"
)

// Alertmanager 0.4 silence format
type silence struct {
	ID       int `json:"id"`
	Matchers []struct {
		Name    string `json:"name"`
		Value   string `json:"value"`
		IsRegex bool   `json:"isRegex"`
	} `json:"matchers"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`
	Comment   string    `json:"comment"`
}

// silenceAPIResponseV04 is what Alertmanager 0.4 API returns
type silenceAPISchema struct {
	Status string `json:"status"`
	Data   struct {
		Silences      []silence `json:"silences"`
		TotalSilences int       `json:"totalSilences"`
	} `json:"data"`
	Error string `json:"error"`
}

// SilenceMapper implements Alertmanager 0.4 API schema
type SilenceMapper struct {
	mapper.SilenceMapper
}

// IsSupported returns true if given version string is supported
func (m SilenceMapper) IsSupported(version string) bool {
	versionRange := semver.MustParseRange(">=0.4.0 <0.5.0")
	return versionRange(semver.MustParse(version))
}

func (m SilenceMapper) Decode(source io.ReadCloser) ([]models.Silence, error) {
	silences := []models.Silence{}
	resp := silenceAPISchema{}

	defer source.Close()
	err := json.NewDecoder(source).Decode(resp)
	if err != nil {
		return silences, err
	}

	if resp.Status != "success" {
		return silences, errors.New(resp.Error)
	}

	for _, s := range resp.Data.Silences {
		us := models.Silence{
			ID:        strconv.Itoa(s.ID),
			Matchers:  s.Matchers,
			StartsAt:  s.StartsAt,
			EndsAt:    s.EndsAt,
			CreatedAt: s.CreatedAt,
			CreatedBy: s.CreatedBy,
			Comment:   s.Comment,
		}
		silences = append(silences, us)
	}
	return silences, nil
}
