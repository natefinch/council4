// Package corpuscheck provides deterministic parallel checks over Scryfall
// card bulk-data arrays.
package corpuscheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// CheckFunc checks one root or face Oracle text.
type CheckFunc func(Text) []Issue

// Text identifies one Oracle-text field in Scryfall bulk data.
type Text struct {
	Index           int
	FaceOrder       int
	ID              string
	OracleID        string
	Name            string
	FaceName        string
	Set             string
	CollectorNumber string
	OracleText      string
	TypeLine        string
}

// Issue is one localized lexer or parser problem.
type Issue struct {
	Severity string      `json:"severity,omitempty"`
	Reason   string      `json:"reason"`
	Detail   string      `json:"detail,omitempty"`
	Text     string      `json:"text,omitempty"`
	Span     shared.Span `json:"span"`
}

// UnsupportedCard is one root or face text with issues.
type UnsupportedCard struct {
	ID              string  `json:"id"`
	OracleID        string  `json:"oracle_id,omitempty"`
	Name            string  `json:"name"`
	FaceName        string  `json:"face_name,omitempty"`
	Set             string  `json:"set,omitempty"`
	CollectorNumber string  `json:"collector_number,omitempty"`
	OracleText      string  `json:"oracle_text"`
	Issues          []Issue `json:"issues"`
}

// Report summarizes one full corpus check.
type Report struct {
	CardCount        int               `json:"card_count"`
	OracleTextCount  int               `json:"oracle_text_count"`
	UnsupportedCount int               `json:"unsupported_count"`
	Unsupported      []UnsupportedCard `json:"unsupported"`
}

type scryfallCard struct {
	ID              string         `json:"id"`
	OracleID        string         `json:"oracle_id"`
	Name            string         `json:"name"`
	Set             string         `json:"set"`
	CollectorNumber string         `json:"collector_number"`
	OracleText      string         `json:"oracle_text"`
	TypeLine        string         `json:"type_line"`
	CardFaces       []scryfallFace `json:"card_faces"`
}

type scryfallFace struct {
	OracleID   string `json:"oracle_id"`
	Name       string `json:"name"`
	OracleText string `json:"oracle_text"`
	TypeLine   string `json:"type_line"`
}

type checkedText struct {
	UnsupportedCard

	index     int
	faceOrder int
}

// Check stream-decodes a Scryfall card array and checks its Oracle texts with a
// bounded worker pool.
func Check(input io.Reader, workers int, check CheckFunc) (Report, error) {
	if workers < 1 {
		return Report{}, errors.New("workers must be at least 1")
	}
	decoder := json.NewDecoder(input)
	token, err := decoder.Token()
	if err != nil {
		return Report{}, fmt.Errorf("reading card array: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return Report{}, errors.New("scryfall bulk data must be a JSON array")
	}

	jobs := make(chan Text)
	results := make(chan checkedText)
	var workerGroup sync.WaitGroup
	for range workers {
		workerGroup.Go(func() {
			for text := range jobs {
				issues := check(text)
				if len(issues) == 0 {
					continue
				}
				results <- checkedText{
					index:     text.Index,
					faceOrder: text.FaceOrder,
					UnsupportedCard: UnsupportedCard{
						ID:              text.ID,
						OracleID:        text.OracleID,
						Name:            text.Name,
						FaceName:        text.FaceName,
						Set:             text.Set,
						CollectorNumber: text.CollectorNumber,
						OracleText:      text.OracleText,
						Issues:          issues,
					},
				}
			}
		})
	}

	var collected []checkedText
	var collector sync.WaitGroup
	collector.Go(func() {
		for result := range results {
			collected = append(collected, result)
		}
	})

	var cards int
	var texts int
	for decoder.More() {
		var card scryfallCard
		if err := decoder.Decode(&card); err != nil {
			close(jobs)
			workerGroup.Wait()
			close(results)
			collector.Wait()
			return Report{}, fmt.Errorf("decoding card %d: %w", cards, err)
		}
		index := cards
		cards++
		if card.OracleText != "" {
			jobs <- textFromCard(index, 0, card, scryfallFace{
				OracleID:   card.OracleID,
				Name:       card.Name,
				OracleText: card.OracleText,
				TypeLine:   card.TypeLine,
			}, "")
			texts++
		}
		for faceIndex, face := range card.CardFaces {
			if face.OracleText == "" {
				continue
			}
			jobs <- textFromCard(index, faceIndex+1, card, face, face.Name)
			texts++
		}
	}
	close(jobs)
	workerGroup.Wait()
	close(results)
	collector.Wait()

	if _, err := decoder.Token(); err != nil {
		return Report{}, fmt.Errorf("closing card array: %w", err)
	}
	slices.SortFunc(collected, func(a, b checkedText) int {
		if a.index != b.index {
			return a.index - b.index
		}
		return a.faceOrder - b.faceOrder
	})
	unsupported := make([]UnsupportedCard, len(collected))
	for i := range collected {
		unsupported[i] = collected[i].UnsupportedCard
	}
	return Report{
		CardCount:        cards,
		OracleTextCount:  texts,
		UnsupportedCount: len(unsupported),
		Unsupported:      unsupported,
	}, nil
}

func textFromCard(
	index, faceOrder int,
	card scryfallCard,
	face scryfallFace,
	faceName string,
) Text {
	return Text{
		Index:           index,
		FaceOrder:       faceOrder,
		ID:              card.ID,
		OracleID:        face.OracleID,
		Name:            card.Name,
		FaceName:        faceName,
		Set:             card.Set,
		CollectorNumber: card.CollectorNumber,
		OracleText:      face.OracleText,
		TypeLine:        face.TypeLine,
	}
}

// WriteText writes a terminal-oriented report.
func WriteText(output io.Writer, report Report) error {
	if _, err := fmt.Fprintf(
		output,
		"Checked %d Oracle texts from %d cards; %d unsupported.\n",
		report.OracleTextCount,
		report.CardCount,
		report.UnsupportedCount,
	); err != nil {
		return err
	}
	for _, card := range report.Unsupported {
		name := card.Name
		if card.FaceName != "" {
			name += " / " + card.FaceName
		}
		if _, err := fmt.Fprintf(output, "\n%s (%s %s)\n", name, card.Set, card.CollectorNumber); err != nil {
			return err
		}
		for _, problem := range card.Issues {
			if _, err := fmt.Fprintf(
				output,
				"  %d:%d: %s: %q\n",
				problem.Span.Start.Line,
				problem.Span.Start.Column,
				problem.Reason,
				problem.Text,
			); err != nil {
				return err
			}
		}
	}
	return nil
}
