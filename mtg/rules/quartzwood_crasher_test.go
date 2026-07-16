package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// quartzwoodCrasherPattern is the reusable "Whenever one or more creatures you
// control with trample deal combat damage to a player" trigger pattern, matching
// the compiled Quartzwood Crasher card. It coalesces simultaneous combat damage
// to the same player into one trigger (OneOrMorePerDamagedPlayer) and matches
// only controlled creatures with trample as the damage source.
func quartzwoodCrasherPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                     game.EventDamageDealt,
		Controller:                game.TriggerControllerYou,
		Subject:                   game.TriggerSubjectDamageSource,
		OneOrMore:                 true,
		OneOrMorePerDamagedPlayer: true,
		RequireCombatDamage:       true,
		DamageRecipient:           game.DamageRecipientPlayer,
		DamageSourceSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Keyword:       game.Trample,
		},
	}
}

// quartzwoodCrasherTokenDef is the X/X green Dinosaur Beast with trample the
// ability creates; CreateToken sizes it dynamically from the batch's combat
// damage.
func quartzwoodCrasherTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Dinosaur Beast",
		Colors:   []color.Color{color.Green},
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Dinosaur, types.Beast},
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
	}}
}

// addQuartzwoodCrasher adds the trigger holder plus its create-token instruction
// that sizes the token from the coalesced combat-damage batch.
func addQuartzwoodCrasher(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addTriggeredPermanent(g, controller, quartzwoodCrasherPattern(), []game.Instruction{{
		Primitive: game.CreateToken{
			Amount:    game.Fixed(1),
			Source:    game.TokenDef(quartzwoodCrasherTokenDef()),
			Power:     opt.Val(game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountTriggeringEventTotalCombatDamage, Multiplier: 1})),
			Toughness: opt.Val(game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountTriggeringEventTotalCombatDamage, Multiplier: 1})),
		},
	}}, nil)
}

// trampleAttacker adds a creature with trample controlled by controller, usable
// as a matching combat-damage source for the Quartzwood trigger.
func trampleAttacker(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
	}})
}

// vanillaAttacker adds a creature without trample.
func vanillaAttacker(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
	}})
}

type combatDeal struct {
	source *game.Permanent
	player game.PlayerID
	amount int
	combat bool
}

// dealBatchedCombatDamage deals each source's damage to its player and then
// assigns the shared SimultaneousID the real combat engine would, so all combat
// events from one pass coalesce like a single combat-damage step.
func dealBatchedCombatDamage(g *game.Game, deals ...combatDeal) {
	eventStart := len(g.Events)
	for _, d := range deals {
		dealPlayerDamage(g, d.source.CardInstanceID, d.source.ObjectID, d.source.Controller, d.player, d.amount, d.combat)
	}
	batchCombatDamageEvents(g, eventStart)
}

// resolveAllTriggers puts every pending triggered ability on the stack and
// resolves them, returning how many distinct triggers fired.
func resolveAllTriggers(engine *Engine, g *game.Game) int {
	if !engine.putTriggeredAbilitiesOnStack(g) {
		return 0
	}
	fired := g.Stack.Size()
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	return fired
}

// dinosaurBeastSizes returns the sorted power of every created Dinosaur Beast
// token (power equals toughness for these X/X tokens).
func dinosaurBeastSizes(g *game.Game) []int {
	var sizes []int
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Dinosaur Beast" {
			sizes = append(sizes, effectivePower(g, permanent))
		}
	}
	slices.Sort(sizes)
	return sizes
}

// TestQuartzwoodCrasherBatchesSamePlayerCombatDamage proves two trample
// creatures dealing combat damage to the same player coalesce into one trigger
// whose X is the sum they dealt.
func TestQuartzwoodCrasherBatchesSamePlayerCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addQuartzwoodCrasher(g, game.Player1)
	three := trampleAttacker(g, game.Player1, "Three")
	four := trampleAttacker(g, game.Player1, "Four")

	dealBatchedCombatDamage(g,
		combatDeal{three, game.Player2, 3, true},
		combatDeal{four, game.Player2, 4, true},
	)

	if fired := resolveAllTriggers(engine, g); fired != 1 {
		t.Fatalf("triggers fired = %d, want 1 (one coalesced trigger for the shared player)", fired)
	}

	if sizes := dinosaurBeastSizes(g); len(sizes) != 1 || sizes[0] != 7 {
		t.Fatalf("Dinosaur Beast sizes = %v, want one 7/7 token (3+4 combat damage)", sizes)
	}

}

func TestQuartzwoodCrasherUsesEventTimeTrample(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addQuartzwoodCrasher(g, game.Player1)
	attacker := trampleAttacker(g, game.Player1, "Transient Trample")
	dealBatchedCombatDamage(g, combatDeal{attacker, game.Player2, 5, true})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:                 g.IDGen.Next(),
		AffectedObjectID:   attacker.ObjectID,
		Layer:              game.LayerAbility,
		RemoveAllAbilities: true,
		Duration:           game.DurationPermanent,
	})
	if fired := resolveAllTriggers(engine, g); fired != 1 {
		t.Fatalf("triggers fired = %d, want 1 from event-time trample", fired)
	}
	if sizes := dinosaurBeastSizes(g); len(sizes) != 1 || sizes[0] != 5 {
		t.Fatalf("Dinosaur Beast sizes = %v, want one 5/5 token", sizes)
	}
}

// TestQuartzwoodCrasherSeparateTriggersPerDamagedPlayer proves combat damage to
// different players produces separate triggers, each sized by its own player's
// damage, even though the combat engine shares one SimultaneousID.
func TestQuartzwoodCrasherSeparateTriggersPerDamagedPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addQuartzwoodCrasher(g, game.Player1)
	toTwoA := trampleAttacker(g, game.Player1, "ToTwoA")
	toTwoB := trampleAttacker(g, game.Player1, "ToTwoB")
	toThree := trampleAttacker(g, game.Player1, "ToThree")

	dealBatchedCombatDamage(g,
		combatDeal{toTwoA, game.Player2, 5, true},
		combatDeal{toTwoB, game.Player2, 1, true},
		combatDeal{toThree, game.Player3, 2, true},
	)

	if fired := resolveAllTriggers(engine, g); fired != 2 {
		t.Fatalf("triggers fired = %d, want 2 (one per damaged player)", fired)
	}
	if sizes := dinosaurBeastSizes(g); len(sizes) != 2 || sizes[0] != 2 || sizes[1] != 6 {
		t.Fatalf("Dinosaur Beast sizes = %v, want [2 6] (Player2 got 5+1, Player3 got 2)", sizes)
	}
}

// TestQuartzwoodCrasherIgnoresNonMatchingDamage proves only matching combat
// damage from controlled trample creatures counts: non-trample creatures and
// non-combat damage neither trigger nor add to X.
func TestQuartzwoodCrasherIgnoresNonMatchingDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addQuartzwoodCrasher(g, game.Player1)
	trample := trampleAttacker(g, game.Player1, "Trampler")
	nonTrample := vanillaAttacker(g, game.Player1, "Grounded")
	noncombat := trampleAttacker(g, game.Player1, "Pinger")

	dealBatchedCombatDamage(g,
		combatDeal{trample, game.Player2, 4, true},
		combatDeal{nonTrample, game.Player2, 3, true},
		combatDeal{noncombat, game.Player2, 2, false},
	)

	if fired := resolveAllTriggers(engine, g); fired != 1 {
		t.Fatalf("triggers fired = %d, want 1 (only the trample combat source triggers)", fired)
	}
	if sizes := dinosaurBeastSizes(g); len(sizes) != 1 || sizes[0] != 4 {
		t.Fatalf("Dinosaur Beast sizes = %v, want one 4/4 token (only trample combat damage counts)", sizes)
	}
}

// TestQuartzwoodCrasherUsesActualDamageDealt proves X reads the damage actually
// dealt after prevention/replacement, not the attackers' raw power: a prevention
// shield that reduces one attacker's damage lowers X accordingly, and a fully
// prevented attacker contributes nothing (no event emitted).
func TestQuartzwoodCrasherUsesActualDamageDealt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addQuartzwoodCrasher(g, game.Player1)
	partial := trampleAttacker(g, game.Player1, "Partial")
	fully := trampleAttacker(g, game.Player1, "Fully")

	// Player2 prevents the next 2 damage from Partial and all damage from Fully.
	g.PreventionShields = append(g.PreventionShields,
		game.PreventionShield{
			ID:                g.IDGen.Next(),
			Controller:        game.Player2,
			Player:            game.Player2,
			SourcePermanentID: partial.ObjectID,
			Amount:            2,
		},
		game.PreventionShield{
			ID:                g.IDGen.Next(),
			Controller:        game.Player2,
			Player:            game.Player2,
			SourcePermanentID: fully.ObjectID,
			All:               true,
		},
	)

	dealBatchedCombatDamage(g,
		combatDeal{partial, game.Player2, 5, true},
		combatDeal{fully, game.Player2, 6, true},
	)

	if fired := resolveAllTriggers(engine, g); fired != 1 {
		t.Fatalf("triggers fired = %d, want 1 (only the partially prevented attacker deals damage)", fired)
	}
	if sizes := dinosaurBeastSizes(g); len(sizes) != 1 || sizes[0] != 3 {
		t.Fatalf("Dinosaur Beast sizes = %v, want one 3/3 token (5-2 dealt, fully prevented adds nothing)", sizes)
	}
}

// TestQuartzwoodCrasherSourceLeavesBeforeResolution proves X still reflects the
// combat damage the batch dealt after the damaging creatures leave the
// battlefield before the trigger resolves, because X reads the recorded damage
// events, not the (now absent) sources.
func TestQuartzwoodCrasherSourceLeavesBeforeResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addQuartzwoodCrasher(g, game.Player1)
	three := trampleAttacker(g, game.Player1, "Three")
	four := trampleAttacker(g, game.Player1, "Four")

	dealBatchedCombatDamage(g,
		combatDeal{three, game.Player2, 3, true},
		combatDeal{four, game.Player2, 4, true},
	)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger was not put on stack")
	}
	// Both damaging creatures leave before the trigger resolves.
	if _, ok := destroyPermanent(g, three.ObjectID); !ok {
		t.Fatal("failed to destroy first attacker")
	}
	if _, ok := destroyPermanent(g, four.ObjectID); !ok {
		t.Fatal("failed to destroy second attacker")
	}
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	if sizes := dinosaurBeastSizes(g); len(sizes) != 1 || sizes[0] != 7 {
		t.Fatalf("Dinosaur Beast sizes = %v, want one 7/7 token (X reads recorded damage after sources leave)", sizes)
	}
}
