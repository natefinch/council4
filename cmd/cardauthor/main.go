// Command cardauthor creates unregistered authored-Go drafts, not card behavior.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type oracleCard struct {
	ID            string   `json:"id"`
	OracleID      string   `json:"oracle_id"`
	Name          string   `json:"name"`
	Layout        string   `json:"layout"`
	ManaCost      string   `json:"mana_cost"`
	TypeLine      string   `json:"type_line"`
	OracleText    string   `json:"oracle_text"`
	Colors        []string `json:"colors"`
	ColorIdentity []string `json:"color_identity"`
	Power         string   `json:"power"`
	Toughness     string   `json:"toughness"`
}

func main() {
	corpus := flag.String("corpus", ".cardwork/corpus/oracle-cards.json", "pinned Oracle JSON")
	oracleID := flag.String("oracle-id", "", "exact Oracle identity")
	symbol := flag.String("symbol", "", "exported Go entry identifier")
	out := flag.String("out", "", "new scratch directory; never production")
	flag.Parse()
	if err := run(*corpus, *oracleID, *symbol, *out); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(corpus, oracleID, symbol, out string) error {
	if oracleID == "" || out == "" || !token.IsIdentifier(symbol) || !token.IsExported(symbol) {
		return errors.New("require -oracle-id, exported -symbol and new -out directory")
	}
	clean := filepath.Clean(out)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) ||
		!strings.HasPrefix(filepath.ToSlash(clean), ".cardwork/scaffold/") {
		return errors.New("output must be a new directory beneath .cardwork/scaffold")
	}
	data, err := os.ReadFile(corpus)
	if err != nil {
		return err
	}
	var input []oracleCard
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	var found []oracleCard
	for _, card := range input {
		if card.OracleID == oracleID {
			found = append(found, card)
		}
	}
	if len(found) != 1 {
		return fmt.Errorf("oracle identity must resolve exactly once, got %d", len(found))
	}
	body, scenario, err := render(found[0], symbol)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(clean), 0o700); err != nil {
		return err
	}
	if err := os.Mkdir(clean, 0o700); err != nil {
		return err
	}
	for name, content := range map[string][]byte{"card.go": body, "scenario_test.go": scenario} {
		if err := os.WriteFile(filepath.Join(clean, name), content, 0o600); err != nil {
			return err
		}
	}
	fmt.Printf("Unregistered draft for %s (%s), revision %s: %s\n", found[0].Name, oracleID, found[0].ID, clean)
	return nil
}

func render(card oracleCard, symbol string) (body, scenario []byte, err error) {
	if card.Layout != "normal" {
		return nil, nil, fmt.Errorf("scaffold only accepts normal single-face cards; got %q", card.Layout)
	}
	metadata, err := metadataLiteral(card, symbol)
	if err != nil {
		return nil, nil, err
	}
	body, err = format.Source(fmt.Appendf(nil, `package authored

import (
 "fmt"
 "github.com/natefinch/council4/mtg/cardprogram"
 "github.com/natefinch/council4/mtg/cards/cardset"
 "github.com/natefinch/council4/mtg/game"
 "github.com/natefinch/council4/mtg/game/color"
 "github.com/natefinch/council4/mtg/game/cost"
 "github.com/natefinch/council4/mtg/game/types"
 "github.com/natefinch/council4/opt"
)

// Oracle input revision: %s. Draft only; do not register until complete.
var %[2]sProgram = game.ProgramReference{ID: %[3]q + "/spell/0", Version: "1"}
var %[2]s = cardset.Entry{
 OracleID: %[3]q, Name: %[4]q, New: new%[2]s,
 Programs: []cardprogram.Program{{Reference: %[2]sProgram, Resolve: resolve%[2]s}},
}

func new%[2]s() *game.CardDef {
 return &game.CardDef{
  %[5]s
 }
}

func resolve%[2]s(ctx cardprogram.Context) error {
 // Example operation only. Replace with the entire card's actual behavior,
 // declare its targets/costs/trigger as needed, and remove the unsupported error.
 if _, err := ctx.Execute(game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}); err != nil {
  return err
 }
 return fmt.Errorf("unsupported draft: %[2]s is not implemented")
}
`, card.ID, symbol, card.OracleID, card.Name, metadata))
	if err != nil {
		return nil, nil, err
	}
	scenario, err = format.Source(fmt.Appendf(nil, `package rules

import (
 "testing"
 "github.com/natefinch/council4/mtg/cardprogram"
 "github.com/natefinch/council4/mtg/cards/authored"
 "github.com/natefinch/council4/mtg/game"
)

func TestAuthored%[1]sScenario(t *testing.T) {
 catalog, err := cardprogram.NewCatalog(authored.%[1]s.Programs)
 if err != nil { t.Fatal(err) }
 engine := NewEngineWithPrograms(nil, catalog)
 g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
 // Add real targets, choices, replacements and source departures appropriate
 // to the card. For a permanent/trigger use the ordinary entry/event harness.
 addImplementationSpellToStack(g, game.Player1, authored.%[1]s.New(), nil)
 engine.resolveTopOfStack(g, &TurnLog{})
 t.Fatal("replace this assertion with the complete named-card behavior checks")
}
`, symbol))
	return body, scenario, err
}

func metadataLiteral(card oracleCard, symbol string) (string, error) {
	colors := map[string]string{"W": "White", "U": "Blue", "B": "Black", "R": "Red", "G": "Green"}
	colorList := func(values []string) (string, error) {
		var result []string
		for _, value := range values {
			name, ok := colors[value]
			if !ok {
				return "", fmt.Errorf("unknown color %q", value)
			}
			result = append(result, "color."+name)
		}
		return strings.Join(result, ","), nil
	}
	printed, err := colorList(card.Colors)
	if err != nil {
		return "", err
	}
	identity, err := colorList(card.ColorIdentity)
	if err != nil {
		return "", err
	}
	var symbols []string
	remaining := card.ManaCost
	for remaining != "" {
		if !strings.HasPrefix(remaining, "{") || !strings.Contains(remaining, "}") {
			return "", errors.New("malformed mana metadata")
		}
		end := strings.IndexByte(remaining, '}')
		value := remaining[1:end]
		remaining = remaining[end+1:]
		if len(value) == 1 && strings.ContainsRune("WUBRGCXS", rune(value[0])) {
			symbols = append(symbols, "cost."+value)
		} else if n, err := strconv.Atoi(value); err == nil && n >= 0 {
			symbols = append(symbols, fmt.Sprintf("cost.O(%d)", n))
		} else {
			return "", fmt.Errorf("scaffold needs ordinary single-symbol/numeric mana metadata, got %q; author this metadata directly", value)
		}
	}
	var supers, cardTypes, subs []string
	parts := strings.Split(card.TypeLine, " — ")
	for word := range strings.FieldsSeq(parts[0]) {
		if word == "Basic" || word == "Legendary" || word == "Snow" || word == "World" {
			supers = append(supers, strconv.Quote(word))
		} else {
			cardTypes = append(cardTypes, strconv.Quote(word))
		}
	}
	if len(parts) > 1 {
		for word := range strings.FieldsSeq(parts[1]) {
			subs = append(subs, strconv.Quote(word))
		}
	}
	var pt strings.Builder
	for _, value := range []struct{ field, number string }{{"Power", card.Power}, {"Toughness", card.Toughness}} {
		if value.number == "" {
			continue
		}
		n, err := strconv.Atoi(value.number)
		if err != nil {
			return "", fmt.Errorf("author variable %s metadata directly", value.field)
		}
		_, _ = fmt.Fprintf(&pt, "%s: opt.Val(game.PT{Value: %d}),\n", value.field, n)
	}
	return fmt.Sprintf(`ColorIdentity: color.NewIdentity(%s),
CardFace: game.CardFace{
 Name: %q, OracleText: %q,
 ManaCost: opt.Val(cost.Mana{%s}),
 Supertypes: []types.Super{%s}, Types: []types.Card{%s}, Subtypes: []types.Sub{%s},
 Colors: []color.Color{%s}, %s
 SpellAbility: opt.Val(game.Mode{Text: "TODO: complete declarations", Program: %sProgram}.Ability()),
},`, identity, card.Name, card.OracleText, strings.Join(symbols, ","), strings.Join(supers, ","), strings.Join(cardTypes, ","), strings.Join(subs, ","), printed, pt.String(), symbol), nil
}
