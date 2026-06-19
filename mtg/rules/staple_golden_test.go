package rules

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	cardb "github.com/natefinch/council4/mtg/cards/b"
	cardc "github.com/natefinch/council4/mtg/cards/c"
	cardl "github.com/natefinch/council4/mtg/cards/l"
	cardr "github.com/natefinch/council4/mtg/cards/r"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// updateStapleGolden rewrites the golden files under testdata/staples when set.
// Run: go test ./mtg/rules/ -run TestStaple -update.
var updateStapleGolden = flag.Bool("update", false, "update the staple golden files in testdata")

// These golden tests run a handful of iconic, supported Commander staples
// through the full real-card pipeline — cast the actual registry card (paying
// its real cost and choosing real targets), resolve it, and settle state-based
// actions — then snapshot the key outcomes (life totals, where each tracked
// card ended up, and what tokens exist). The golden files pin those outcomes so
// behavioral drift in any of these staples shows up as a diff.
//
// To add another staple: pick a supported card from the registry, add a case to
// stapleCases with a builder that stages the board/hand/mana and returns the
// cast to make (target matcher) plus the cards to track, then run the test once
// with -update to record its golden file.

// stapleCreature is a vanilla creature used as a removal/wipe target.
func stapleCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

// stapleScenario starts a scenario whose turn state allows both instant- and
// sorcery-speed casts by Player1 (active player, main phase, empty stack).
func stapleScenario(t *testing.T) *scenario {
	s := newScenario(t)
	s.g.Turn.ActivePlayer = game.Player1
	s.g.Turn.PriorityPlayer = game.Player1
	s.g.Turn.Phase = game.PhasePrecombatMain
	s.g.Turn.Step = game.StepNone
	return s
}

// addMana adds the given amount of a color to a player's pool.
func addMana(s *scenario, player game.PlayerID, color mana.Color, amount int) {
	s.g.Players[player].ManaPool.Add(color, amount)
}

// stapleCase is one staple: a human label and a builder that stages the game
// and returns the cast to make and the cards to report on.
type stapleCase struct {
	name  string
	build func(t *testing.T, s *scenario) stapleSetup
}

// stapleSetup is what a case builder hands back: the caster, the card to cast,
// a matcher selecting the desired cast (by its targets), the agents that answer
// any resolution choices, and the cards to track in the snapshot.
type stapleSetup struct {
	caster  game.PlayerID
	cardID  id.ID
	match   func(action.CastSpellAction) bool
	agents  [game.NumPlayers]PlayerAgent
	tracked []trackedCard
}

type trackedCard struct {
	label string
	id    id.ID
}

func stapleCases() []stapleCase {
	return []stapleCase{
		{name: "lightning-bolt-kills-creature", build: buildBoltKillsCreature},
		{name: "lightning-bolt-burns-opponent", build: buildBoltBurnsOpponent},
		{name: "beast-within-destroys-and-makes-token", build: buildBeastWithin},
		{name: "chandras-ignition-board-wipe", build: buildChandrasIgnition},
		{name: "rampant-growth-ramps", build: buildRampantGrowth},
	}
}

func buildBoltKillsCreature(t *testing.T, s *scenario) stapleSetup {
	t.Helper()
	bear := s.permanent(game.Player2, stapleCreature("Grizzly Bears", 2, 2))
	boltID := s.hand(game.Player1, cardl.LightningBolt)
	addMana(s, game.Player1, mana.R, 1)
	target := bear.permanent().ObjectID
	return stapleSetup{
		caster: game.Player1,
		cardID: boltID,
		match:  targetsPermanent(target),
		agents: allPassAgents(),
		tracked: []trackedCard{
			{label: "Lightning Bolt", id: boltID},
			{label: "Grizzly Bears", id: bear.permanent().CardInstanceID},
		},
	}
}

func buildBoltBurnsOpponent(t *testing.T, s *scenario) stapleSetup {
	t.Helper()
	boltID := s.hand(game.Player1, cardl.LightningBolt)
	addMana(s, game.Player1, mana.R, 1)
	return stapleSetup{
		caster:  game.Player1,
		cardID:  boltID,
		match:   targetsPlayer(game.Player2),
		agents:  allPassAgents(),
		tracked: []trackedCard{{label: "Lightning Bolt", id: boltID}},
	}
}

func buildBeastWithin(t *testing.T, s *scenario) stapleSetup {
	t.Helper()
	land := s.permanent(game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	spellID := s.hand(game.Player1, cardb.BeastWithin)
	addMana(s, game.Player1, mana.G, 1)
	addMana(s, game.Player1, mana.C, 2)
	target := land.permanent().ObjectID
	return stapleSetup{
		caster: game.Player1,
		cardID: spellID,
		match:  targetsPermanent(target),
		agents: allPassAgents(),
		tracked: []trackedCard{
			{label: "Beast Within", id: spellID},
			{label: "Forest", id: land.permanent().CardInstanceID},
		},
	}
}

func buildChandrasIgnition(t *testing.T, s *scenario) stapleSetup {
	t.Helper()
	bomber := s.permanent(game.Player1, stapleCreature("Hellkite", 5, 5))
	smallA := s.permanent(game.Player2, stapleCreature("Goblin", 1, 1))
	smallB := s.permanent(game.Player3, stapleCreature("Soldier", 2, 2))
	spellID := s.hand(game.Player1, cardc.ChandraSIgnition)
	addMana(s, game.Player1, mana.R, 2)
	addMana(s, game.Player1, mana.C, 3)
	target := bomber.permanent().ObjectID
	return stapleSetup{
		caster: game.Player1,
		cardID: spellID,
		match:  targetsPermanent(target),
		agents: allPassAgents(),
		tracked: []trackedCard{
			{label: "Chandra's Ignition", id: spellID},
			{label: "Hellkite", id: bomber.permanent().CardInstanceID},
			{label: "Goblin", id: smallA.permanent().CardInstanceID},
			{label: "Soldier", id: smallB.permanent().CardInstanceID},
		},
	}
}

func buildRampantGrowth(t *testing.T, s *scenario) stapleSetup {
	t.Helper()
	forestID := s.library(game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Forest},
	}})
	spellID := s.hand(game.Player1, cardr.RampantGrowth)
	addMana(s, game.Player1, mana.G, 1)
	addMana(s, game.Player1, mana.C, 1)
	agents := allPassAgents()
	agents[game.Player1] = &stapleSearchAgent{wanted: "Forest"}
	return stapleSetup{
		caster: game.Player1,
		cardID: spellID,
		match:  noTargets,
		agents: agents,
		tracked: []trackedCard{
			{label: "Rampant Growth", id: spellID},
			{label: "Forest", id: forestID},
		},
	}
}

// targetsPermanent matches a cast whose target list contains the given permanent.
func targetsPermanent(objectID id.ID) func(action.CastSpellAction) bool {
	return func(cast action.CastSpellAction) bool {
		return slices.ContainsFunc(cast.Targets, func(target game.Target) bool {
			return target.Kind == game.TargetPermanent && target.PermanentID == objectID
		})
	}
}

// targetsPlayer matches a cast whose target list contains the given player.
func targetsPlayer(player game.PlayerID) func(action.CastSpellAction) bool {
	return func(cast action.CastSpellAction) bool {
		return slices.ContainsFunc(cast.Targets, func(target game.Target) bool {
			return target.Kind == game.TargetPlayer && target.PlayerID == player
		})
	}
}

// noTargets matches a cast that targets nothing.
func noTargets(cast action.CastSpellAction) bool {
	return len(cast.Targets) == 0
}

func allPassAgents() [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for i := range agents {
		agents[i] = staplePassAgent{}
	}
	return agents
}

type staplePassAgent struct{}

func (staplePassAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

// stapleSearchAgent answers a search by selecting the option whose label matches
// the wanted card name.
type stapleSearchAgent struct {
	wanted string
}

func (*stapleSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *stapleSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		if option.Label == a.wanted {
			return []int{option.Index}
		}
	}
	return nil
}

func TestStapleGoldenOutcomes(t *testing.T) {
	for _, sc := range stapleCases() {
		t.Run(sc.name, func(t *testing.T) {
			s := stapleScenario(t)
			setup := sc.build(t, s)

			cast, ok := findStapleCast(s, setup)
			if !ok {
				t.Fatal("no legal cast of the staple matched the requested targets")
			}
			if !s.engine().applyActionWithChoices(s.game(), setup.caster, cast, setup.agents, &TurnLog{}) {
				t.Fatal("applying the staple cast failed")
			}
			s.engine().resolveTopOfStackWithChoices(s.game(), setup.agents, &TurnLog{})
			s.engine().applyStateBasedActions(s.game())

			got := renderStapleOutcome(s.game(), setup.tracked)
			checkStapleGolden(t, sc.name+".txt", got)
		})
	}
}

// TestStapleGoldenIsDeterministic guards against unstable ordering: rendering
// the same resolved state twice must be byte-identical.
func TestStapleGoldenIsDeterministic(t *testing.T) {
	for _, sc := range stapleCases() {
		t.Run(sc.name, func(t *testing.T) {
			first := runStaple(t, sc)
			second := runStaple(t, sc)
			if !bytes.Equal(first, second) {
				t.Errorf("staple %s render is not deterministic", sc.name)
			}
		})
	}
}

func runStaple(t *testing.T, sc stapleCase) []byte {
	t.Helper()
	s := stapleScenario(t)
	setup := sc.build(t, s)
	cast, ok := findStapleCast(s, setup)
	if !ok {
		t.Fatal("no legal cast of the staple matched the requested targets")
	}
	s.engine().applyActionWithChoices(s.game(), setup.caster, cast, setup.agents, &TurnLog{})
	s.engine().resolveTopOfStackWithChoices(s.game(), setup.agents, &TurnLog{})
	s.engine().applyStateBasedActions(s.game())
	return renderStapleOutcome(s.game(), setup.tracked)
}

func findStapleCast(s *scenario, setup stapleSetup) (action.Action, bool) {
	for _, act := range s.legalActions(setup.caster) {
		if act.Kind != action.ActionCastSpell {
			continue
		}
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != setup.cardID {
			continue
		}
		if setup.match(cast) {
			return act, true
		}
	}
	return action.Action{}, false
}

// renderStapleOutcome produces a deterministic snapshot of the post-resolution
// state: life totals, where each tracked card is, and the tokens on the board.
func renderStapleOutcome(g *game.Game, tracked []trackedCard) []byte {
	var b strings.Builder
	_, _ = fmt.Fprintln(&b, "life:")
	for player := range g.Players {
		_, _ = fmt.Fprintf(&b, "  Player%d: %d\n", player+1, g.Players[player].Life)
	}
	_, _ = fmt.Fprintln(&b, "tracked cards:")
	for _, card := range tracked {
		_, _ = fmt.Fprintf(&b, "  %s: %s\n", card.label, locateCard(g, card.id))
	}
	tokens := boardTokens(g)
	_, _ = fmt.Fprintln(&b, "tokens on battlefield:")
	if len(tokens) == 0 {
		_, _ = fmt.Fprintln(&b, "  (none)")
	}
	for _, token := range tokens {
		_, _ = fmt.Fprintf(&b, "  %s\n", token)
	}
	return []byte(b.String())
}

// locateCard reports the zone (and owner) holding a card instance.
func locateCard(g *game.Game, cardID id.ID) string {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return fmt.Sprintf("battlefield (controller Player%d)", permanent.Controller+1)
		}
	}
	for player := range g.Players {
		p := g.Players[player]
		switch {
		case p.Hand.Contains(cardID):
			return fmt.Sprintf("hand (Player%d)", player+1)
		case p.Graveyard.Contains(cardID):
			return fmt.Sprintf("graveyard (Player%d)", player+1)
		case p.Library.Contains(cardID):
			return fmt.Sprintf("library (Player%d)", player+1)
		case p.Exile.Contains(cardID):
			return fmt.Sprintf("exile (Player%d)", player+1)
		}
	}
	return "gone"
}

// boardTokens lists the token permanents on the battlefield as
// "<name> (controller PlayerN)", sorted for a stable snapshot.
func boardTokens(g *game.Game) []string {
	var tokens []string
	for _, permanent := range g.Battlefield {
		if permanent.TokenDef == nil {
			continue
		}
		tokens = append(tokens, fmt.Sprintf("%s (controller Player%d)", permanent.TokenDef.Name, permanent.Controller+1))
	}
	slices.Sort(tokens)
	return tokens
}

func checkStapleGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", "staples", name)
	if *updateStapleGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatalf("update golden %s: %v", name, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run with -update to create): %v", name, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("staple golden %s drift.\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}
