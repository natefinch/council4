package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ascendSourceDef is a creature carrying the permanent ascend static ability
// (CR 702.131b), the shape cardgen lowers Wayward Swordtooth / Snubhorn Sentry's
// "Ascend" keyword into.
func ascendSourceDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{game.AscendStaticBody},
	}}
}

func vanillaDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
	}}
}

// addVanillaPermanents adds n plain permanents controlled by the given player.
func addVanillaPermanents(g *game.Game, controller game.PlayerID, n int) {
	for range n {
		addCombatPermanent(g, controller, vanillaDef("Filler"))
	}
}

// TestAscendGrantsAtTenPermanents covers CR 702.131b: a player controlling an
// ascend permanent gets the city's blessing once they control ten or more
// permanents, and not before.
func TestAscendGrantsAtTenPermanents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 8) // source + 8 = 9 permanents

	if checkAscendCityBlessing(g) {
		t.Fatal("granted the city's blessing at nine permanents")
	}
	if g.Players[game.Player1].HasCityBlessing {
		t.Fatal("player has the city's blessing at nine permanents")
	}

	addVanillaPermanents(g, game.Player1, 1) // now 10 permanents

	if !checkAscendCityBlessing(g) {
		t.Fatal("did not grant the city's blessing at ten permanents")
	}
	if !g.Players[game.Player1].HasCityBlessing {
		t.Fatal("player did not get the city's blessing at ten permanents")
	}
	if got := countEvents(g, game.EventGotCityBlessing); got != 1 {
		t.Fatalf("EventGotCityBlessing count = %d, want 1", got)
	}
}

// TestAscendPersistsAfterSourceLeaves covers the CR 702.131 ruling that the
// city's blessing is kept for the rest of the game even after the ascend source
// leaves and the controller drops below ten permanents.
func TestAscendPersistsAfterSourceLeaves(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 9) // source + 9 = 10 permanents

	if !checkAscendCityBlessing(g) {
		t.Fatal("did not grant the city's blessing at ten permanents")
	}

	// The ascend source and most permanents leave; the player drops well below
	// ten permanents.
	g.Battlefield = g.Battlefield[:0]
	_ = source

	if changed := checkAscendCityBlessing(g); changed {
		t.Fatal("city's blessing was re-granted or toggled after the source left")
	}
	if !g.Players[game.Player1].HasCityBlessing {
		t.Fatal("city's blessing was lost after the source left")
	}
}

// TestAscendNoGrantIfSourceLeavesBeforeTen proves a player who never reaches ten
// permanents while controlling the ascend source never gets the city's blessing,
// even after the source leaves.
func TestAscendNoGrantIfSourceLeavesBeforeTen(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 5) // source + 5 = 6 permanents

	if checkAscendCityBlessing(g) {
		t.Fatal("granted the city's blessing below ten permanents")
	}

	g.Battlefield = g.Battlefield[:0]

	if checkAscendCityBlessing(g) {
		t.Fatal("granted the city's blessing after the source left")
	}
	if g.Players[game.Player1].HasCityBlessing {
		t.Fatal("player got the city's blessing without ever reaching ten permanents")
	}
}

// TestAscendCountsTokens confirms token permanents count toward the ten-permanent
// threshold (CR 702.131 ruling: tokens count).
func TestAscendCountsTokens(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	for range 9 {
		addTokenCreaturePermanent(g, game.Player1, "Saproling")
	}

	if !checkAscendCityBlessing(g) {
		t.Fatal("tokens did not count toward the ten-permanent threshold")
	}
	if !g.Players[game.Player1].HasCityBlessing {
		t.Fatal("player did not get the city's blessing with token permanents")
	}
}

// TestAscendIgnoresPhasedOutPermanents covers the CR 702.131 ruling that
// phased-out permanents are not counted.
func TestAscendIgnoresPhasedOutPermanents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 8) // source + 8 = 9 active permanents
	phased := addCombatPermanent(g, game.Player1, vanillaDef("Blinked"))
	phased.PhasedOut = true // tenth permanent is phased out

	if checkAscendCityBlessing(g) {
		t.Fatal("a phased-out permanent was counted toward the threshold")
	}

	phased.PhasedOut = false // phases back in -> ten active permanents

	if !checkAscendCityBlessing(g) {
		t.Fatal("did not grant the city's blessing once the permanent phased in")
	}
}

// TestAscendIgnoresOpponentPermanents confirms only permanents the ascend
// controller controls count (CR 702.131b: "you control ten or more permanents").
func TestAscendIgnoresOpponentPermanents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 4)  // controller has five permanents total
	addVanillaPermanents(g, game.Player2, 20) // opponent's board is large

	if checkAscendCityBlessing(g) {
		t.Fatal("opponent permanents were counted toward the controller's threshold")
	}
	if g.Players[game.Player1].HasCityBlessing {
		t.Fatal("player got the city's blessing from opponent permanents")
	}
}

// TestAscendEachPlayerIndependent covers CR 702.131c: the city's blessing is a
// per-player designation, so each player gets it independently.
func TestAscendEachPlayerIndependent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 9) // Player1: 10 permanents

	addCombatPermanent(g, game.Player2, ascendSourceDef("Snubhorn Sentry"))
	addVanillaPermanents(g, game.Player2, 4) // Player2: 5 permanents

	checkAscendCityBlessing(g)

	if !g.Players[game.Player1].HasCityBlessing {
		t.Fatal("Player1 did not get the city's blessing at ten permanents")
	}
	if g.Players[game.Player2].HasCityBlessing {
		t.Fatal("Player2 got the city's blessing below ten permanents")
	}

	addVanillaPermanents(g, game.Player2, 5) // Player2 now has 10 permanents
	checkAscendCityBlessing(g)

	if !g.Players[game.Player2].HasCityBlessing {
		t.Fatal("Player2 did not get the city's blessing once at ten permanents")
	}
}

// TestAscendNoDuplicateEventOnRepeatedChecks proves the grant is idempotent:
// re-running the continuous check emits no second EventGotCityBlessing and
// reports no change once the player already has the blessing.
func TestAscendNoDuplicateEventOnRepeatedChecks(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 9)

	if !checkAscendCityBlessing(g) {
		t.Fatal("first check did not grant the city's blessing")
	}
	if changed := checkAscendCityBlessing(g); changed {
		t.Fatal("second check reported a change for an already-blessed player")
	}
	if got := countEvents(g, game.EventGotCityBlessing); got != 1 {
		t.Fatalf("EventGotCityBlessing count = %d, want 1", got)
	}
}

// TestAscendSpellGrantsOnResolve covers the spell form of ascend (CR 702.131a):
// the GainCityBlessing primitive grants the resolving controller the city's
// blessing when they control ten or more permanents, and not otherwise.
func TestAscendSpellGrantsOnResolve(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addVanillaPermanents(g, game.Player1, 9) // below ten permanents

	obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.GainCityBlessing{}, &TurnLog{})

	if g.Players[game.Player1].HasCityBlessing {
		t.Fatal("spell ascend granted the city's blessing below ten permanents")
	}

	addVanillaPermanents(g, game.Player1, 1) // now ten permanents
	resolveInstruction(engine, g, obj, game.GainCityBlessing{}, &TurnLog{})

	if !g.Players[game.Player1].HasCityBlessing {
		t.Fatal("spell ascend did not grant the city's blessing at ten permanents")
	}
	if got := countEvents(g, game.EventGotCityBlessing); got != 1 {
		t.Fatalf("EventGotCityBlessing count = %d, want 1", got)
	}
}

// TestAscendGrantedDuringStateBasedActions proves the permanent ascend check
// runs as part of the state-based-action loop (CR 702.131b is a continuous
// check, not a triggered ability), so the city's blessing is granted when SBAs
// are processed with ten or more permanents in play.
func TestAscendGrantedDuringStateBasedActions(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, ascendSourceDef("Wayward Swordtooth"))
	addVanillaPermanents(g, game.Player1, 9) // source + 9 = 10 permanents

	engine.applyStateBasedActions(g)

	if !g.Players[game.Player1].HasCityBlessing {
		t.Fatal("state-based actions did not grant the city's blessing at ten permanents")
	}
	if got := countEvents(g, game.EventGotCityBlessing); got != 1 {
		t.Fatalf("EventGotCityBlessing count = %d, want 1", got)
	}
}

// cityBlessingCombatGuardDef builds the "can't attack or block unless you have
// the city's blessing" static (Wayward Swordtooth), the shape cardgen lowers
// that guard into: source-scoped can't-attack / can't-block rule effects gated
// on the negated city's-blessing designation.
func cityBlessingCombatGuardDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Text:      name + " can't attack or block unless you have the city's blessing.",
			Condition: opt.Val(game.Condition{ControllerHasCityBlessing: true, Negate: true}),
			RuleEffects: []game.RuleEffect{
				{Kind: game.RuleEffectCantAttack, AffectedSource: true},
				{Kind: game.RuleEffectCantBlock, AffectedSource: true},
			},
		}},
	}}
}

func sourceHasCantAttack(g *game.Game, source *game.Permanent) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectCantAttack && effect.AffectedObjectID == source.ObjectID {
			return true
		}
	}
	return false
}

// TestAscendCombatGuardRespectsCityBlessing proves the "can't attack or block
// unless you have the city's blessing" guard (Wayward Swordtooth) prohibits
// combat while the controller lacks the blessing and lifts once the controller
// has it.
func TestAscendCombatGuardRespectsCityBlessing(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, cityBlessingCombatGuardDef("Wayward Swordtooth"))

	if !sourceHasCantAttack(g, source) {
		t.Fatal("combat guard did not prohibit attacking without the city's blessing")
	}

	g.Players[game.Player1].HasCityBlessing = true

	if sourceHasCantAttack(g, source) {
		t.Fatal("combat guard still prohibited attacking with the city's blessing")
	}
}
