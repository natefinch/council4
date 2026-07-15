package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// phyrexianBeastTokenDef is the 4/4 green Phyrexian Beast token Ezuri's Predation
// creates, the fighting subject of the correlated fight.
func phyrexianBeastTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Phyrexian Beast",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Beast},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	}}
}

// ezuriPredationInstructions is the lowered Ezuri's Predation spell: create one
// 4/4 token per opponent-controlled creature (publishing both the tokens and the
// counted creatures) then fight each token against a distinct counted creature.
func ezuriPredationInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.CreateToken{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountCountSelector,
				Multiplier: 1,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerOpponent,
				}),
			}),
			Source:            game.TokenDef(phyrexianBeastTokenDef()),
			PublishLinked:     game.LinkedKey("correlated-fight-tokens"),
			PublishCountGroup: game.LinkedKey("correlated-fight-creatures"),
		}},
		{Primitive: game.CorrelatedFight{
			Subjects: "correlated-fight-tokens",
			Objects:  "correlated-fight-creatures",
		}},
	}
}

func resolveEzuriPredation(g *game.Game) {
	engine := NewEngine(nil)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Ezuri's Predation"}}),
		Controller: game.Player1,
	}
	instrs := ezuriPredationInstructions()
	log := TurnLog{}
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
}

// TestCorrelatedFightMultipleOpponents proves the base one-to-one shape across
// multiple opponents: three opponent creatures (two under Player2, one under
// Player3) yield three 4/4 tokens, and each counted creature takes 4 damage from
// a distinct token while each token takes its creature's power in return.
func TestCorrelatedFightMultipleOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	c1 := addCreatureWithPowerToughness(g, game.Player2, 2, 2)
	c2 := addCreatureWithPowerToughness(g, game.Player2, 3, 3)
	c3 := addCreatureWithPowerToughness(g, game.Player3, 1, 5)

	resolveEzuriPredation(g)

	if got := countTokensNamed(g, "Phyrexian Beast", game.Player1); got != 3 {
		t.Fatalf("created tokens = %d, want 3", got)
	}
	for _, want := range []struct {
		permanent *game.Permanent
		damage    int
	}{{c1, 4}, {c2, 4}, {c3, 4}} {
		if got := want.permanent.MarkedDamage; got != want.damage {
			t.Fatalf("counted creature marked damage = %d, want %d", got, want.damage)
		}
	}
	totalTokenDamage := 0
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Token {
			totalTokenDamage += permanent.MarkedDamage
		}
	}
	// Each creature dealt its power to its paired token: 2 + 3 + 1 = 6.
	if totalTokenDamage != 6 {
		t.Fatalf("total token marked damage = %d, want 6", totalTokenDamage)
	}
}

// TestCorrelatedFightPartialDeparture proves fight-time legality: a counted
// creature that leaves the battlefield between the create and the fight is
// re-resolved by object ID, found gone, and its pair skipped, while the surviving
// creatures still fight. The departed creature's token deals and takes no damage.
func TestCorrelatedFightPartialDeparture(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	surviving := addCreatureWithPowerToughness(g, game.Player2, 2, 2)
	leaving := addCreatureWithPowerToughness(g, game.Player2, 3, 3)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Ezuri's Predation"}}),
		Controller: game.Player1,
	}
	instrs := ezuriPredationInstructions()
	log := TurnLog{}
	// Resolve only the create instruction first, so both creatures are counted
	// and two tokens are published.
	engine.resolveInstructionWithChoices(g, obj, &instrs[0], [game.NumPlayers]PlayerAgent{}, &log)
	if got := countTokensNamed(g, "Phyrexian Beast", game.Player1); got != 2 {
		t.Fatalf("created tokens = %d, want 2", got)
	}
	// Remove one counted creature before the fight resolves.
	if _, ok := removePermanentFromBattlefield(g, leaving.ObjectID); !ok {
		t.Fatal("failed to remove counted creature before fight")
	}

	engine.resolveInstructionWithChoices(g, obj, &instrs[1], [game.NumPlayers]PlayerAgent{}, &log)

	if got := surviving.MarkedDamage; got != 4 {
		t.Fatalf("surviving creature marked damage = %d, want 4", got)
	}
	// Exactly one token took damage (from the surviving creature, power 2); the
	// token paired with the departed creature never fought.
	tokensDamaged := 0
	totalTokenDamage := 0
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Token && permanent.MarkedDamage > 0 {
			tokensDamaged++
			totalTokenDamage += permanent.MarkedDamage
		}
	}
	if tokensDamaged != 1 || totalTokenDamage != 2 {
		t.Fatalf("tokens damaged = %d total = %d, want 1 token taking 2", tokensDamaged, totalTokenDamage)
	}
}

// TestCorrelatedFightTokenDoublerNoExtraFights proves that a token doubler does
// not create extra fight pairs: with Primal Vigor doubling creation, two counted
// creatures produce four tokens, yet only the two counted creatures are fought
// (min of four tokens and two creatures), matching the official ruling that the
// surplus tokens do not fight.
func TestCorrelatedFightTokenDoublerNoExtraFights(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyControllerTokenDoublingCardDef())
	c1 := addCreatureWithPowerToughness(g, game.Player2, 2, 2)
	c2 := addCreatureWithPowerToughness(g, game.Player3, 1, 1)

	resolveEzuriPredation(g)

	if got := countTokensNamed(g, "Phyrexian Beast", game.Player1); got != 4 {
		t.Fatalf("created tokens = %d, want 4 (doubled)", got)
	}
	if got := c1.MarkedDamage; got != 4 {
		t.Fatalf("creature c1 marked damage = %d, want 4", got)
	}
	if got := c2.MarkedDamage; got != 4 {
		t.Fatalf("creature c2 marked damage = %d, want 4", got)
	}
	// Only two fights occurred, so only two of the four tokens took damage
	// (creature powers 2 + 1 = 3 total).
	tokensDamaged := 0
	totalTokenDamage := 0
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Token && permanent.MarkedDamage > 0 {
			tokensDamaged++
			totalTokenDamage += permanent.MarkedDamage
		}
	}
	if tokensDamaged != 2 || totalTokenDamage != 3 {
		t.Fatalf("tokens damaged = %d total = %d, want 2 tokens taking 3 total", tokensDamaged, totalTokenDamage)
	}
}

// TestCorrelatedFightNoCreatures proves the empty case: with no opponent
// creatures, no tokens are created and no fights occur, and the sequence resolves
// without error.
func TestCorrelatedFightNoCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	resolveEzuriPredation(g)

	if got := countTokensNamed(g, "Phyrexian Beast", game.Player1); got != 0 {
		t.Fatalf("created tokens = %d, want 0", got)
	}
}
