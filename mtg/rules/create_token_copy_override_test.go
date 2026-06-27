package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// createOverrideCopyToken resolves a single copy of the given source permanent
// under the supplied override spec and returns the created token's definition.
func createOverrideCopyToken(t *testing.T, g *game.Game, spec game.TokenCopySpec) *game.CardDef {
	t.Helper()
	target := g.Battlefield[0]
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	resolved := handleCreateToken(r, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(spec),
	})
	if !resolved.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil {
			return permanent.TokenDef
		}
	}
	t.Fatal("no copy token created")
	return nil
}

// TestCreateCopyTokenAdditiveOverrides verifies the additive copy-token
// characteristic overrides ("except it's a 3/3 black Zombie artifact in addition
// to its other colors and types") append to the copied permanent's colors, card
// types, and subtypes while the power/toughness override replaces them.
func TestCreateCopyTokenAdditiveOverrides(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	def := createOverrideCopyToken(t, g, game.TokenCopySpec{
		Source:       game.TokenCopySourceObject,
		Object:       game.TargetPermanentReference(0),
		SetPower:     opt.Val(game.PT{Value: 3}),
		SetToughness: opt.Val(game.PT{Value: 3}),
		AddColors:    []color.Color{color.Black},
		AddTypes:     []types.Card{types.Artifact},
		AddSubtypes:  []types.Sub{types.Zombie},
	})

	if !slices.Contains(def.Colors, color.Green) || !slices.Contains(def.Colors, color.Black) {
		t.Errorf("colors = %v, want both green and black", def.Colors)
	}
	if !slices.Contains(def.Types, types.Creature) || !slices.Contains(def.Types, types.Artifact) {
		t.Errorf("types = %v, want both Creature and Artifact", def.Types)
	}
	if !slices.Contains(def.Subtypes, types.Bear) || !slices.Contains(def.Subtypes, types.Zombie) {
		t.Errorf("subtypes = %v, want both Bear and Zombie", def.Subtypes)
	}
	if def.Power.Val.Value != 3 || def.Toughness.Val.Value != 3 {
		t.Errorf("P/T = %d/%d, want 3/3", def.Power.Val.Value, def.Toughness.Val.Value)
	}
}

// TestCreateCopyTokenReplacementOverrides verifies the replacing copy-token
// characteristic overrides ("except it's a 1/1 blue Frog") replace the copied
// permanent's colors and subtypes rather than appending to them.
func TestCreateCopyTokenReplacementOverrides(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	def := createOverrideCopyToken(t, g, game.TokenCopySpec{
		Source:       game.TokenCopySourceObject,
		Object:       game.TargetPermanentReference(0),
		SetPower:     opt.Val(game.PT{Value: 1}),
		SetToughness: opt.Val(game.PT{Value: 1}),
		SetColors:    []color.Color{color.Blue},
		SetSubtypes:  []types.Sub{types.Frog},
	})

	if len(def.Colors) != 1 || def.Colors[0] != color.Blue {
		t.Errorf("colors = %v, want [blue]", def.Colors)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Frog {
		t.Errorf("subtypes = %v, want [Frog]", def.Subtypes)
	}
	if def.Power.Val.Value != 1 || def.Toughness.Val.Value != 1 {
		t.Errorf("P/T = %d/%d, want 1/1", def.Power.Val.Value, def.Toughness.Val.Value)
	}
}
