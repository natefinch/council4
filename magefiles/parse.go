package magefiles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// Parse parses a single card's Oracle text with the oracle parser library and
// prints the resulting parser.Document as indented JSON to stdout. The card is
// located by name in the cached Scryfall oracle-cards corpus (overridable with
// the COUNCIL4_ORACLE_CARDS environment variable). Multi-face cards emit one
// document per face with Oracle text.
func Parse(_ context.Context, cardName string) error {
	if strings.TrimSpace(cardName) == "" {
		return errors.New("a card name argument is required")
	}
	corpusPath, err := oracleCardsCachePath()
	if err != nil {
		return err
	}
	file, err := os.Open(corpusPath)
	if err != nil {
		return fmt.Errorf("opening oracle cards corpus %s: %w", corpusPath, err)
	}
	defer file.Close()

	faces, err := findCardFaces(file, cardName)
	if err != nil {
		return err
	}
	if len(faces) == 0 {
		return fmt.Errorf("no card named %q with Oracle text found in %s", cardName, corpusPath)
	}

	results := make([]parsedCardFace, 0, len(faces))
	for _, face := range faces {
		document, _ := parser.Parse(face.oracleText, parser.Context{
			InstantOrSorcery: hasCardType(face.typeLine, "Instant") || hasCardType(face.typeLine, "Sorcery"),
			Planeswalker:     hasCardType(face.typeLine, "Planeswalker"),
			Saga:             hasSubtype(face.typeLine, "Saga"),
			CardName:         face.contextName,
		})
		results = append(results, parsedCardFace{Name: face.displayName, Document: document})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	var payload any = results
	if len(results) == 1 {
		payload = results[0]
	}
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("encoding parse result: %w", err)
	}
	return nil
}

type parsedCardFace struct {
	Name     string          `json:"name"`
	Document parser.Document `json:"document"`
}

type scryfallCardJSON struct {
	Name       string             `json:"name"`
	TypeLine   string             `json:"type_line"`
	OracleText string             `json:"oracle_text"`
	CardFaces  []scryfallFaceJSON `json:"card_faces"`
}

type scryfallFaceJSON struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	OracleText string `json:"oracle_text"`
}

type cardFaceText struct {
	displayName string
	contextName string
	typeLine    string
	oracleText  string
}

// findCardFaces stream-decodes the Scryfall card array and returns the Oracle
// texts of the first card whose card name or face name matches cardName.
func findCardFaces(input io.Reader, cardName string) ([]cardFaceText, error) {
	decoder := json.NewDecoder(input)
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("reading card array: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("oracle cards corpus must be a JSON array")
	}
	for decoder.More() {
		var card scryfallCardJSON
		if err := decoder.Decode(&card); err != nil {
			return nil, fmt.Errorf("decoding card: %w", err)
		}
		if faces := cardFacesIfMatch(card, cardName); faces != nil {
			return faces, nil
		}
	}
	return nil, nil
}

// cardFacesIfMatch returns the Oracle texts to parse when cardName matches the
// card. A full card-name match emits every face with Oracle text; a face-name
// match emits only that face.
func cardFacesIfMatch(card scryfallCardJSON, cardName string) []cardFaceText {
	if strings.EqualFold(card.Name, cardName) {
		var faces []cardFaceText
		if card.OracleText != "" {
			faces = append(faces, cardFaceText{
				displayName: card.Name,
				contextName: "",
				typeLine:    card.TypeLine,
				oracleText:  card.OracleText,
			})
		}
		for _, face := range card.CardFaces {
			if face.OracleText == "" {
				continue
			}
			faces = append(faces, cardFaceText{
				displayName: face.Name,
				contextName: face.Name,
				typeLine:    face.TypeLine,
				oracleText:  face.OracleText,
			})
		}
		return faces
	}
	for _, face := range card.CardFaces {
		if strings.EqualFold(face.Name, cardName) && face.OracleText != "" {
			return []cardFaceText{{
				displayName: face.Name,
				contextName: face.Name,
				typeLine:    face.TypeLine,
				oracleText:  face.OracleText,
			}}
		}
	}
	return nil
}

func hasCardType(typeLine, wanted string) bool {
	mainType, _, _ := strings.Cut(typeLine, "—")
	for word := range strings.FieldsSeq(mainType) {
		if word == wanted {
			return true
		}
	}
	return false
}

func hasSubtype(typeLine, wanted string) bool {
	_, subtypes, ok := strings.Cut(typeLine, "—")
	if !ok {
		return false
	}
	for word := range strings.FieldsSeq(subtypes) {
		if word == wanted {
			return true
		}
	}
	return false
}
