package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// painlandColoredManaAbility builds the colored mana ability of a painland-style
// source: "{T}: Add {color}. <CARDNAME> deals amount damage to you." The lowered
// content is an AddMana instruction followed by a source-dealt Damage to the
// controller, matching the cardgen output for these cards.
func painlandColoredManaAbility(color mana.Color, amount int) game.ManaAbility {
	return game.ManaAbility{
		Text:            "{T}: Add mana. This permanent deals damage to you.",
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: color}},
			{Primitive: game.Damage{
				Amount:       game.Fixed(amount),
				Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			}},
		}}.Ability(),
	}
}

// TestManaAbilitySelfDamageRiderDealsDamage verifies that activating a mana
// ability carrying a "deals N damage to you" rider resolves immediately, taps
// the source, adds the mana, and deals the rider damage to the controller.
func TestManaAbilitySelfDamageRiderDealsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := painlandColoredManaAbility(mana.B, 1)
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Painful Elf", Types: []types.Card{types.Creature}}},
		&body,
	)
	startLife := g.Players[game.Player1].Life
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(self-damage mana ability) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("source was not tapped after activation")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 1 {
		t.Fatalf("mana pool (B) = %d, want 1", got)
	}
	if got := g.Players[game.Player1].Life; got != startLife-1 {
		t.Fatalf("life = %d, want %d after self-damage rider", got, startLife-1)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}

// TestManaAbilitySelfDamageRiderUsesRiderAmount verifies that the rider deals
// the exact fixed amount printed (Ancient Tomb's two, Tarnished Citadel's three)
// rather than a single point.
func TestManaAbilitySelfDamageRiderUsesRiderAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := painlandColoredManaAbility(mana.C, 2)
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Ancient Pain", Types: []types.Card{types.Land}}},
		&body,
	)
	startLife := g.Players[game.Player1].Life
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(two-damage rider mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 1 {
		t.Fatalf("mana pool (C) = %d, want 1", got)
	}
	if got := g.Players[game.Player1].Life; got != startLife-2 {
		t.Fatalf("life = %d, want %d after dealing 2 damage to controller", got, startLife-2)
	}
}

// TestPainlandColorChoiceManaAbilityDealsDamage exercises the full painland
// colored ability shape (Choose a color, add it, then deal the rider damage),
// matching the cardgen output for Shivan Reef and its allies. It resolves
// immediately, adds exactly one of the offered colors, and deals the rider
// damage to the controller.
func TestPainlandColorChoiceManaAbilityDealsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.ManaAbility{
		Text:            "{T}: Add {U} or {R}. This land deals 1 damage to you.",
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Choose{
				Choice: game.ResolutionChoice{
					Kind:   game.ResolutionChoiceMana,
					Prompt: "Choose a color",
					Colors: []mana.Color{mana.U, mana.R},
				},
				PublishChoice: game.ChoiceKey("oracle-mana-color"),
			}},
			{Primitive: game.AddMana{
				Amount:     game.Fixed(1),
				ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
			}},
			{Primitive: game.Damage{
				Amount:       game.Fixed(1),
				Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			}},
		}}.Ability(),
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Shivan Proxy", Types: []types.Card{types.Land}}},
		&body,
	)
	startLife := g.Players[game.Player1].Life
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(painland color-choice ability) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("source was not tapped after activation")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U) + g.Players[game.Player1].ManaPool.Amount(mana.R); got != 1 {
		t.Fatalf("mana pool (U+R) = %d, want 1", got)
	}
	if got := g.Players[game.Player1].Life; got != startLife-1 {
		t.Fatalf("life = %d, want %d after self-damage rider", got, startLife-1)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}

// of a painland: an ordinary tap-for-mana ability must not touch the
// controller's life total.
func TestPlainManaAbilityDealsNoSelfDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.TapManaAbility(mana.C)
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Plain Land", Types: []types.Card{types.Land}}},
		&body,
	)
	startLife := g.Players[game.Player1].Life
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(plain mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].Life; got != startLife {
		t.Fatalf("life = %d, want %d (plain mana ability must deal no damage)", got, startLife)
	}
}
