package rules

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// These property-based tests exercise core engine invariants over many
// seeded pseudo-random games and board states. They use math/rand/v2 with
// explicit seeds (not Go fuzzing) so the whole file stays well under the CI
// budget (a few seconds) while still covering hundreds of distinct states.
//
// Every helper here is prefixed with "prop" so it never collides with the
// shared helpers in the other package rules test files.

// propRand returns a deterministic RNG seeded from a single integer so each
// iteration is independently reproducible from its seed.
func propRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed*0x9E3779B97F4A7C15+1))
}

// --- Property 1 & 2 & 3: full-game invariants ----------------------------

// propRandomAgent picks a uniformly random action from the engine-provided
// legal actions. Because it holds the live *game.Game, it can record
// per-priority invariant violations (a card in two places, or an illegal
// target offered) at every decision point, not just at game end.
//
// Violations are appended to a sink rather than reported through *testing.T
// directly, so a game may run inside a goroutine (under a termination
// deadline) without touching test state concurrently. When sink is nil the
// agent performs no checks.
type propRandomAgent struct {
	seat game.PlayerID
	rng  *rand.Rand
	g    *game.Game
	sink *propViolationSink
}

// propViolationSink collects invariant-violation messages observed while a
// game runs. It is written only from the single goroutine running that game,
// and read by the test goroutine after the game completes (a channel receive
// establishes the happens-before relationship).
//
// cloneBudget caps how many apply-on-clone target validations the whole game
// may perform (shared across all four seats), so the deep-clone cost stays
// bounded regardless of how many casts the agents are offered.
type propViolationSink struct {
	msgs        []string
	cloneBudget int
}

// propNewSink returns a sink with the per-game clone-validation budget.
func propNewSink() *propViolationSink {
	return &propViolationSink{cloneBudget: 6}
}

func (s *propViolationSink) addf(format string, args ...any) {
	s.msgs = append(s.msgs, fmt.Sprintf(format, args...))
}

func (a *propRandomAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	if a.sink != nil {
		a.propCheckInvariants(legal)
	}
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[a.rng.IntN(len(legal))]
}

// propCheckInvariants records the two per-state properties that must hold at
// every priority point.
//
// Property 2 (target legality): an offered cast action is independently
// replayed on a deep clone of the game through the engine's real apply
// pipeline (target resolution, segmentation, payment). The engine must ACCEPT
// it — return true without panicking. This is independent of the generator's
// own target gate, so it genuinely catches the engine offering an
// uncastable/illegally-targeted action. It is bounded to at most two casts per
// priority point and a per-game total (sink.cloneBudget) to keep clone cost low.
//
// Property 1 (no duplication): no real card is ever in two places at once,
// checked at every priority point.
func (a *propRandomAgent) propCheckInvariants(legal []action.Action) {
	const castsPerPoint = 2
	validated := 0
	for _, act := range legal {
		if validated >= castsPerPoint || a.sink.cloneBudget <= 0 {
			break
		}
		if act.Kind != action.ActionCastSpell {
			continue
		}
		cast, ok := act.CastSpellPayload()
		if !ok {
			continue
		}
		validated++
		a.sink.cloneBudget--
		a.propValidateCastOnClone(act, cast)
	}

	if dup := propDuplicatedCards(a.g); len(dup) > 0 {
		a.sink.addf("card(s) present in more than one place: %v", propLocationReport(a.g, dup))
	}
}

// propValidateCastOnClone applies an offered cast on a throwaway deep clone via
// the engine's real apply path and records a violation if the engine rejects
// it or panics. A panic or a false return means the engine offered an action
// it cannot actually carry out (for example, with an illegal target).
func (a *propRandomAgent) propValidateCastOnClone(act action.Action, cast action.CastSpellAction) {
	clone := a.g.Clone()
	accepted, panicked := a.propApplyRecovered(clone, act)
	if panicked {
		a.sink.addf("seat %v: engine offered cast of %q (targets %+v) that panicked when applied to a clone",
			a.seat, a.propCastName(cast), cast.Targets)
		return
	}
	if !accepted {
		a.sink.addf("seat %v: engine offered cast of %q (targets %+v) that applyAction rejected on a clone",
			a.seat, a.propCastName(cast), cast.Targets)
	}
}

// propApplyRecovered applies act on clone, recovering any panic so a genuinely
// illegal offered action surfaces as a violation rather than crashing the test.
func (a *propRandomAgent) propApplyRecovered(clone *game.Game, act action.Action) (accepted, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	// Validate on a throwaway engine driven by the clone's own RNG so a cast
	// that happens to consume randomness can never perturb the live game's
	// engine RNG stream.
	accepted = NewEngine(clone.RNG).applyAction(clone, a.seat, act)
	return accepted, false
}

func (a *propRandomAgent) propCastName(cast action.CastSpellAction) string {
	if card, ok := a.g.GetCardInstance(cast.CardID); ok {
		return cardFaceOrDefault(card, cast.Face).Name
	}
	return "?"
}

// propCountLocations counts, for every real (non-token) CardInstance, how many
// places it currently occupies. It allocates no per-card strings, so it is
// cheap enough to call at every priority point. Tokens never appear because
// they have no CardInstance entry — they live only as permanents with a
// TokenDef.
func propCountLocations(g *game.Game) map[id.ID]int {
	counts := make(map[id.ID]int, len(g.CardInstances))
	for cardID := range g.CardInstances {
		counts[cardID] = 0
	}

	mark := func(cardID id.ID) {
		if _, ok := counts[cardID]; ok {
			counts[cardID]++
		}
	}
	markZone := func(z interface{ Range(func(id.ID) bool) }) {
		z.Range(func(cardID id.ID) bool {
			mark(cardID)
			return true
		})
	}
	for i := range g.Players {
		p := g.Players[i]
		markZone(&p.Library)
		markZone(&p.Hand)
		markZone(&p.Graveyard)
		markZone(&p.Exile)
		markZone(&p.CommandZone)
	}

	for _, perm := range g.Battlefield {
		if perm == nil || perm.Token {
			continue
		}
		mark(perm.CardInstanceID)
		for _, merged := range perm.MergedCards {
			mark(merged.CardInstanceID)
		}
	}

	for _, obj := range g.Stack.Objects() {
		// Only a card-backed spell carries its card on the stack. Abilities
		// keep their source on the battlefield (counted above), and token
		// spell copies have no CardInstance.
		if obj.Kind != game.StackSpell || obj.SourceTokenDef != nil {
			continue
		}
		mark(obj.SourceID)
	}

	return counts
}

// propDuplicatedCards returns the IDs of real cards occupying more than one
// place. This is the always-true half of conservation: nothing the engine
// does should ever clone a card into two zones at once.
func propDuplicatedCards(g *game.Game) []id.ID {
	var dup []id.ID
	for cardID, count := range propCountLocations(g) {
		if count > 1 {
			dup = append(dup, cardID)
		}
	}
	slices.Sort(dup)
	return dup
}

// propMissingCards returns the IDs of real cards that occupy no place at all.
func propMissingCards(g *game.Game) []id.ID {
	var missing []id.ID
	for cardID, count := range propCountLocations(g) {
		if count == 0 {
			missing = append(missing, cardID)
		}
	}
	slices.Sort(missing)
	return missing
}

// propCardPlaces returns the human-readable places a single card occupies. It
// is used only when building a failure report, so the string formatting cost
// is paid only on the rare failure path.
func propCardPlaces(g *game.Game, cardID id.ID) []string {
	var places []string
	for i := range g.Players {
		p := g.Players[i]
		seat := game.PlayerID(i)
		for _, z := range []struct {
			name string
			zone interface{ Contains(id.ID) bool }
		}{
			{"Library", &p.Library}, {"Hand", &p.Hand}, {"Graveyard", &p.Graveyard},
			{"Exile", &p.Exile}, {"Command", &p.CommandZone},
		} {
			if z.zone.Contains(cardID) {
				places = append(places, fmt.Sprintf("P%d.%s", seat, z.name))
			}
		}
	}
	for _, perm := range g.Battlefield {
		if perm == nil || perm.Token {
			continue
		}
		if perm.CardInstanceID == cardID {
			places = append(places, "Battlefield")
		}
		for _, merged := range perm.MergedCards {
			if merged.CardInstanceID == cardID {
				places = append(places, "Battlefield(merged)")
			}
		}
	}
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackSpell && obj.SourceTokenDef == nil && obj.SourceID == cardID {
			places = append(places, "Stack")
		}
	}
	return places
}

func propLocationReport(g *game.Game, cards []id.ID) string {
	var b strings.Builder
	for _, cardID := range cards {
		name := ""
		if inst, ok := g.GetCardInstance(cardID); ok && inst.Def != nil {
			name = inst.Def.Name
		}
		_, _ = fmt.Fprintf(&b, " [%v %q -> %v]", cardID, name, propCardPlaces(g, cardID))
	}
	return b.String()
}

// propRandomAgents builds one random agent per seat sharing the live game and
// violation sink. When sink is nil the agents perform no per-state checks.
func propRandomAgents(seed uint64, g *game.Game, sink *propViolationSink) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for seat := range agents {
		agents[seat] = &propRandomAgent{
			seat: game.PlayerID(seat),
			rng:  propRand(seed*131 + uint64(seat) + 1),
			g:    g,
			sink: sink,
		}
	}
	return agents
}

// TestPropZoneMovesConserveCards exercises property 1 over land-only games.
// Land-only decks never use the stack or abilities, so eliminating a player
// can never orphan a card off the stack. That makes the full conservation
// invariant — no duplicates AND no vanishing — provable at game end over
// every seed.
func TestPropZoneMovesConserveCards(t *testing.T) {
	const iterations = 120
	configs := landOnlyConfigs(8)
	for seed := range uint64(iterations) {
		engine := NewEngine(propRand(seed))
		g := engine.NewGame(configs)
		sink := propNewSink()
		engine.RunGame(g, propRandomAgents(seed, g, sink))
		for _, msg := range sink.msgs {
			t.Errorf("seed %d: %s", seed, msg)
		}
		if dup := propDuplicatedCards(g); len(dup) > 0 {
			t.Fatalf("seed %d: cards in two places at game end:%s", seed, propLocationReport(g, dup))
		}
		if missing := propMissingCards(g); len(missing) > 0 {
			t.Fatalf("seed %d: cards vanished by game end:%s", seed, propLocationReport(g, missing))
		}
	}
}

// TestPropRandomGamesPreserveInvariants exercises properties 1, 2, and 3 over
// mixed decks that drive cards through hand -> stack -> battlefield ->
// graveyard. At every priority point the agents assert no card is duplicated
// (property 1) and every offered cast target is legal (property 2). Every
// game terminating with a winner inside maxGameTurns demonstrates the priority
// loop converges (property 3).
//
// We do NOT assert "no vanishing" here: eliminating a player who has a spell
// on the stack legitimately removes that card from the game (CR 800.4a), which
// is correct behavior, not a leak. The duplication half of conservation has no
// such caveat and is asserted throughout and at game end.
func TestPropRandomGamesPreserveInvariants(t *testing.T) {
	const iterations = 80
	configs := propMixedConfigs()
	for seed := range uint64(iterations) {
		engine := NewEngine(propRand(seed))
		g := engine.NewGame(configs)
		sink := propNewSink()
		result := engine.RunGame(g, propRandomAgents(seed, g, sink))
		for _, msg := range sink.msgs {
			t.Errorf("seed %d: %s", seed, msg)
		}
		if dup := propDuplicatedCards(g); len(dup) > 0 {
			t.Fatalf("seed %d: cards in two places at game end:%s", seed, propLocationReport(g, dup))
		}
		if result.TurnCount >= maxGameTurns {
			t.Fatalf("seed %d: game hit the %d-turn cap without terminating", seed, maxGameTurns)
		}
		if !result.HasWinner {
			t.Fatalf("seed %d: game ended without a winner after %d turns", seed, result.TurnCount)
		}
	}
}

// TestPropPriorityLoopTerminates is the hang guard for property 3: it runs each
// random game under a wall-clock deadline so a priority loop that fails to
// converge is reported instead of hanging the suite indefinitely. RunGame's
// own maxGameTurns cap bounds turn count; this bounds wall-clock time.
func TestPropPriorityLoopTerminates(t *testing.T) {
	const iterations = 16
	const perGameBudget = 5 * time.Second
	configs := propMixedConfigs()
	for seed := range uint64(iterations) {
		engine := NewEngine(propRand(seed))
		g := engine.NewGame(configs)
		done := make(chan *GameResult, 1)
		// nil sink: the goroutine must not touch shared test state.
		agents := propRandomAgents(seed, g, nil)
		go func() { done <- engine.RunGame(g, agents) }()
		select {
		case result := <-done:
			if result.TurnCount >= maxGameTurns {
				t.Fatalf("seed %d: game hit the %d-turn cap without terminating", seed, maxGameTurns)
			}
		case <-time.After(perGameBudget):
			t.Fatalf("seed %d: priority loop did not terminate within %s", seed, perGameBudget)
		}
	}
}

// --- Property 4: state-based actions reach a fixpoint --------------------

// TestPropStateBasedActionsReachFixpoint exercises property 4 over random,
// often-illegal board states. applyStateBasedActions loops internally until
// stable, so a second immediate call must be a complete no-op: no new losses
// and a byte-identical state signature. That proves SBAs converge in one
// engine call.
func TestPropStateBasedActionsReachFixpoint(t *testing.T) {
	const iterations = 300
	for seed := range uint64(iterations) {
		engine := NewEngine(propRand(seed))
		g := propRandomBoard(propRand(seed + 1_000_000))

		// First call converges internally (it loops until stable).
		engine.applyStateBasedActions(g)
		before := propStateSignature(g)

		// A second immediate call must be a complete no-op.
		losses := engine.applyStateBasedActions(g)
		after := propStateSignature(g)

		if len(losses) != 0 {
			t.Fatalf("seed %d: second SBA pass reported %d new losses, want 0", seed, len(losses))
		}
		if before != after {
			t.Fatalf("seed %d: second SBA pass changed state\nbefore: %s\nafter:  %s", seed, before, after)
		}
	}
}

// propRandomBoard builds a random, often-illegal board state covering the SBA
// triggers called out by the property: zero/negative life, zero-toughness and
// lethally damaged creatures, and duplicate legendaries.
func propRandomBoard(rng *rand.Rand) *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	legendNames := []string{"Legend A", "Legend B", "Legend C"}

	for seat := range g.Players {
		controller := game.PlayerID(seat)
		g.Players[seat].Life = rng.IntN(9) - 2 // -2..6, some lethal

		for range rng.IntN(5) {
			toughness := rng.IntN(4) // 0..3, zero-toughness is lethal
			perm := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
				Name:      "Prop Creature",
				Types:     []types.Card{types.Creature},
				Power:     opt.Val(game.PT{Value: rng.IntN(4)}),
				Toughness: opt.Val(game.PT{Value: toughness}),
			}})
			perm.MarkedDamage = rng.IntN(4) // sometimes >= toughness (lethal)
		}

		// Sometimes stack duplicate legendaries to trip the legend rule.
		if rng.IntN(2) == 0 {
			name := legendNames[rng.IntN(len(legendNames))]
			for range 1 + rng.IntN(2) {
				addLegendaryPermanent(g, controller, name)
			}
		}
	}
	return g
}

// propStateSignature captures the SBA-relevant state so two passes can be
// compared for exact equality.
func propStateSignature(g *game.Game) string {
	var b strings.Builder
	for i := range g.Players {
		p := g.Players[i]
		_, _ = fmt.Fprintf(&b, "P%d{life=%d elim=%t gy=%d ex=%d}", i, p.Life, p.Eliminated, p.Graveyard.Size(), p.Exile.Size())
	}
	objects := make([]string, 0, len(g.Battlefield))
	for _, perm := range g.Battlefield {
		if perm == nil {
			continue
		}
		objects = append(objects, fmt.Sprintf("%v:d%d", perm.ObjectID, perm.MarkedDamage))
	}
	slices.Sort(objects)
	for _, o := range objects {
		_, _ = b.WriteString("|" + o)
	}
	return b.String()
}

// --- shared deck builders ------------------------------------------------

// propMixedConfigs builds four identical small decks of lands, vanilla
// creatures, and a couple of targeted spells. The engine shuffles each
// library from its own RNG, so varying only the engine seed yields a fresh
// game every iteration.
func propMixedConfigs() [game.NumPlayers]game.PlayerConfig {
	var configs [game.NumPlayers]game.PlayerConfig
	for seat := range configs {
		var deck []*game.CardDef
		for range 5 {
			deck = append(deck, basicLand())
		}
		for range 2 {
			deck = append(deck, propVanillaCreature())
		}
		deck = append(deck, propCreatureBolt(), propPlayerBolt())
		configs[seat].Deck = deck
	}
	return configs
}

func propVanillaCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Prop Bear",
		ManaCost:  greenCost(),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func propCreatureBolt() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Prop Creature Bolt",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability()),
	}}
}

func propPlayerBolt() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Prop Player Bolt",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability()),
	}}
}
