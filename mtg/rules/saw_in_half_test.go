package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// sawInHalfInstructions returns the resolved instruction sequence of Saw in Half:
// destroy the target creature, then (gated on the destruction succeeding) its
// controller creates two tokens that are copies of it with power and toughness
// each halved and rounded up.
func sawInHalfInstructions() []game.Instruction {
	const resultKey = game.ResultKey("dies-this-way-copy")
	return []game.Instruction{
		{
			Primitive:     game.Destroy{Object: game.TargetPermanentReference(0)},
			PublishResult: resultKey,
		},
		{
			Primitive: game.CreateToken{
				Amount: game.Fixed(2),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source:                     game.TokenCopySourceObject,
					Object:                     game.TargetPermanentReference(0),
					HalvePowerToughnessRoundUp: true,
				}),
				Recipient: opt.Val(game.AffectedTargetControllerReference(0)),
			},
			ResultGate: opt.Val(game.InstructionResultGate{Key: resultKey, Succeeded: game.TriTrue}),
		},
	}
}

// TestSawInHalfDestroysAndCreatesTwoHalvedCopies proves the full Saw in Half
// sequence: destroying a creature its controller (Player2, distinct from the
// spell's controller Player1) does not control the spell yields two token copies
// under that controller, each with the destroyed creature's power and toughness
// halved and rounded up from its last-known information (5/5 -> 3/3).
func TestSawInHalfDestroysAndCreatesTwoHalvedCopies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ogre",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ogre},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	addInstructionSpellToStackForController(g, game.Player1,
		sawInHalfInstructions(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("target creature was not destroyed")
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("destroyed creature was not in its owner's graveyard")
	}

	var copies []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Ogre" {
			copies = append(copies, permanent)
		}
	}
	if len(copies) != 2 {
		t.Fatalf("token copies created = %d, want 2", len(copies))
	}
	for _, token := range copies {
		if token.Controller != game.Player2 {
			t.Errorf("copy controller = %v, want Player2 (the destroyed creature's controller)", token.Controller)
		}
		if token.TokenDef.Power.Val.Value != 3 || token.TokenDef.Toughness.Val.Value != 3 {
			t.Errorf("copy P/T = %d/%d, want 3/3 (half of 5/5 rounded up)",
				token.TokenDef.Power.Val.Value, token.TokenDef.Toughness.Val.Value)
		}
	}
}

// TestSawInHalfCopyHalvesRoundingUp proves the halved-copy power/toughness
// modifier rounds each characteristic up independently, so an odd stat rounds up
// while an even one halves exactly and the two never borrow from each other.
func TestSawInHalfCopyHalvesRoundingUp(t *testing.T) {
	cases := []struct {
		name                     string
		power, toughness         int
		wantPower, wantToughness int
	}{
		{"odd square 5/5", 5, 5, 3, 3},
		{"even square 4/4", 4, 4, 2, 2},
		{"mixed 7/3", 7, 3, 4, 2},
		{"unit 1/1", 1, 1, 1, 1},
		{"zero power 0/1", 0, 1, 0, 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:      "Source",
				Types:     []types.Card{types.Creature},
				Power:     opt.Val(game.PT{Value: test.power}),
				Toughness: opt.Val(game.PT{Value: test.toughness}),
			}})

			def := createOverrideCopyToken(t, g, game.TokenCopySpec{
				Source:                     game.TokenCopySourceObject,
				Object:                     game.TargetPermanentReference(0),
				HalvePowerToughnessRoundUp: true,
			})

			if def.Power.Val.Value != test.wantPower || def.Toughness.Val.Value != test.wantToughness {
				t.Errorf("halved copy P/T = %d/%d, want %d/%d",
					def.Power.Val.Value, def.Toughness.Val.Value, test.wantPower, test.wantToughness)
			}
		})
	}
}
