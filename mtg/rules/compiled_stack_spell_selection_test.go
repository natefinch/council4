package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledStackSpellSelectionPredicates(t *testing.T) {
	for _, tc := range []struct {
		selector string
		types    []types.Card
		colors   []color.Color
		mv       int
		want     bool
	}{
		{"blue enchantment spell you don't control with mana value 3 or less", []types.Card{types.Enchantment}, []color.Color{color.Blue}, 3, true},
		{"blue enchantment spell you don't control with mana value 3 or less", []types.Card{types.Enchantment}, []color.Color{color.Blue, color.Red}, 2, true},
		{"blue enchantment spell you don't control with mana value 3 or less", []types.Card{types.Artifact}, []color.Color{color.Blue}, 3, false},
		{"blue enchantment spell you don't control with mana value 3 or less", []types.Card{types.Enchantment}, []color.Color{color.Red}, 3, false},
		{"blue enchantment spell you don't control with mana value 3 or less", []types.Card{types.Enchantment}, []color.Color{color.Blue}, 4, false},
		{"nonblue enchantment spell", []types.Card{types.Enchantment}, []color.Color{color.Red}, 3, true},
		{"nonblue enchantment spell", []types.Card{types.Enchantment}, []color.Color{color.Blue, color.Red}, 3, false},
		{"colorless artifact spell", []types.Card{types.Artifact}, nil, 3, true},
		{"colorless artifact spell", []types.Card{types.Artifact}, []color.Color{color.Blue}, 3, false},
		{"multicolored noncreature spell", []types.Card{types.Enchantment}, []color.Color{color.Blue, color.Red}, 3, true},
		{"multicolored noncreature spell", []types.Card{types.Enchantment}, []color.Color{color.Blue}, 3, false},
		{"multicolored noncreature spell", []types.Card{types.Enchantment, types.Creature}, []color.Color{color.Blue, color.Red}, 3, false},
		{"noncreature enchantment spell", []types.Card{types.Enchantment}, nil, 3, true},
		{"noncreature enchantment spell", []types.Card{types.Enchantment, types.Creature}, nil, 3, false},
		{"instant or sorcery spell with mana value 2", []types.Card{types.Instant}, nil, 2, true},
		{"instant or sorcery spell with mana value 2", []types.Card{types.Sorcery}, nil, 2, true},
		{"instant or sorcery spell with mana value 2", []types.Card{types.Enchantment}, nil, 2, false},
		{"instant or sorcery spell with mana value 2", []types.Card{types.Instant}, nil, 3, false},
		{"artifact, enchantment, or battle spell with mana value 3 or greater", []types.Card{types.Battle}, nil, 3, true},
		{"artifact, enchantment, or battle spell with mana value 3 or greater", []types.Card{types.Artifact}, nil, 4, true},
		{"artifact, enchantment, or battle spell with mana value 3 or greater", []types.Card{types.Creature}, nil, 3, false},
		{"artifact, enchantment, or battle spell with mana value 3 or greater", []types.Card{types.Enchantment}, nil, 2, false},
	} {
		t.Run(fmt.Sprintf("%s/%v/%v/%d", tc.selector, tc.types, tc.colors, tc.mv), func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Predicate Probe", Layout: "normal", TypeLine: "Instant",
				OracleText: "Counter target " + tc.selector + ".",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addStackSpellWithFace(g, game.Player2, &game.CardFace{
				Name: "Candidate", Types: tc.types, Colors: tc.colors, ManaCost: opt.Val(cost.Mana{cost.O(tc.mv)}),
			})
			g.CardInstances[target.SourceID].Owner = game.Player1
			spec := &def.SpellAbility.Val.Modes[0].Targets[0]
			if got := stackObjectTargetMatchesSpec(g, game.Player1, 0, spec, target.ID); got != tc.want {
				t.Fatalf("matches = %v, want %v", got, tc.want)
			}
			target.Kind = game.StackActivatedAbility
			if stackObjectTargetMatchesSpec(g, game.Player1, 0, spec, target.ID) {
				t.Fatal("spell selector accepted an activated ability")
			}
			target.Kind = game.StackSpell
			if stackObjectTargetMatchesSpec(g, game.Player1, target.ID, spec, target.ID) {
				t.Fatal("spell selector accepted the resolving source itself")
			}
		})
	}
}

func TestCompiledStackSpellControllerNotOwner(t *testing.T) {
	for _, tc := range []struct {
		suffix     string
		controller game.PlayerID
		owner      game.PlayerID
		want       bool
	}{
		{" you control", game.Player1, game.Player2, true},
		{" you control", game.Player2, game.Player1, false},
		{" an opponent controls", game.Player2, game.Player1, true},
		{" an opponent controls", game.Player1, game.Player2, false},
		{" you don't control", game.Player2, game.Player1, true},
		{" you don't control", game.Player1, game.Player2, false},
	} {
		t.Run(fmt.Sprintf("%s/%d", tc.suffix, tc.controller), func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Controller Probe", Layout: "normal", TypeLine: "Instant",
				OracleText: "Counter target enchantment spell" + tc.suffix + " with mana value 3 or less.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addStackSpellWithFace(g, tc.controller, &game.CardFace{
				Name: "Foreign Owned", Types: []types.Card{types.Enchantment}, ManaCost: opt.Val(cost.Mana{cost.O(3)}),
			})
			g.CardInstances[target.SourceID].Owner = tc.owner
			spec := &def.SpellAbility.Val.Modes[0].Targets[0]
			if got := stackObjectTargetMatchesSpec(g, game.Player1, 0, spec, target.ID); got != tc.want {
				t.Fatalf("matches = %v, want %v (owner %v, controller %v)", got, tc.want, tc.owner, tc.controller)
			}
			pushCounterTaxSpell(g, def, target)
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if _, remains := stackObjectByID(g, target.ID); remains == tc.want {
				t.Fatalf("target remains = %v, want %v", remains, !tc.want)
			}
			if got := g.Players[tc.owner].Graveyard.Contains(target.SourceID); got != tc.want {
				t.Fatalf("owner's graveyard contains target = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompiledStackSpellBestowCharacteristics(t *testing.T) {
	g, engine := setupBestowMain(t)
	target, _ := castBestowTester(t, g, engine)
	for _, tc := range []struct {
		selector string
		want     bool
	}{
		{"enchantment spell you don't control", true},
		{"noncreature enchantment spell you don't control", true},
		{"creature spell you don't control", false},
		{"artifact or enchantment spell you don't control", true},
	} {
		t.Run(tc.selector, func(t *testing.T) {
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Bestow Probe", Layout: "normal", TypeLine: "Instant",
				OracleText: "Counter target " + tc.selector + ".",
			})
			spec := &def.SpellAbility.Val.Modes[0].Targets[0]
			if got := stackObjectTargetMatchesSpec(g, game.Player2, 0, spec, target.ID); got != tc.want {
				t.Fatalf("bestowed spell matches = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompiledStackSpellGroupExile(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Group Probe", Layout: "normal", TypeLine: "Instant",
		OracleText: "Exile any number of target artifact or enchantment spells an opponent controls with mana value 3 or less.",
	})
	for _, count := range []int{0, 1, 2} {
		t.Run(cost.Mana{cost.O(count)}.String(), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addStackSpellWithFace(g, game.Player2, &game.CardFace{
				Name: "First", Types: []types.Card{types.Artifact}, ManaCost: opt.Val(cost.Mana{cost.O(2)}),
			})
			second := addStackSpellWithFace(g, game.Player2, &game.CardFace{
				Name: "Second", Types: []types.Card{types.Enchantment}, ManaCost: opt.Val(cost.Mana{cost.O(3)}),
			})
			g.CardInstances[first.SourceID].Owner = game.Player3
			g.CardInstances[second.SourceID].Owner = game.Player4
			resolving := pushCounterTaxSpell(g, def, first)
			resolving.Targets = []game.Target{game.StackObjectTarget(first.ID), game.StackObjectTarget(second.ID)}[:count]
			resolving.TargetCounts = []int{count}
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			for i, obj := range []*game.StackObject{first, second} {
				wantExile := i < count
				if _, remains := stackObjectByID(g, obj.ID); remains == wantExile {
					t.Fatalf("target %d remains = %v, want %v", i, remains, !wantExile)
				}
				owner := g.CardInstances[obj.SourceID].Owner
				if got := g.Players[owner].Exile.Contains(obj.SourceID); got != wantExile {
					t.Fatalf("target %d owner's exile contains = %v, want %v", i, got, wantExile)
				}
			}
		})
	}
}
