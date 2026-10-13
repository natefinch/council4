package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledPermanentUnion(t *testing.T, text string) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Union Capability", Layout: "normal", TypeLine: "Instant", OracleText: text,
	})
}

func TestCompiledPermanentUnionTargetChoices(t *testing.T) {
	for _, tc := range []struct {
		selector string
		want     []bool
	}{
		{"creature or Vehicle", []bool{true, true, true, false, true}},
		{"blue creature or artifact", []bool{false, true, true, true, true}},
		{"artifact creature or Vehicle", []bool{false, false, true, false, true}},
		{"creature or artifact", []bool{true, true, true, true, true}},
		{"artifact creature", []bool{false, false, false, false, true}},
	} {
		t.Run(tc.selector, func(t *testing.T) {
			def := compiledPermanentUnion(t, "Tap target "+tc.selector+".")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			candidates := []game.CardFace{
				{Name: "Creature", Types: []types.Card{types.Creature}},
				{Name: "Blue Creature", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Blue}},
				{Name: "Vehicle", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Vehicle}},
				{Name: "Other Artifact", Types: []types.Card{types.Artifact}},
				{Name: "Artifact Creature", Types: []types.Card{types.Artifact, types.Creature}},
			}
			result := targetChoicesForSpell(g, game.Player1, def, nil, game.CastBranch{})
			if result.kind == targetLegalChoicesFound {
				t.Fatal("union acquired a target without a permanent")
			}
			targets := make([]game.Target, len(candidates))
			for i, face := range candidates {
				permanent := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: face})
				targets[i] = game.PermanentTarget(permanent.ObjectID)
			}
			result = targetChoicesForSpell(g, game.Player1, def, nil, game.CastBranch{})
			if result.kind != targetLegalChoicesFound {
				t.Fatalf("choices: kind=%v err=%v", result.kind, result.err)
			}
			for i, target := range targets {
				got := slices.ContainsFunc(result.choices, func(choice []game.Target) bool {
					return slices.Contains(choice, target)
				})
				if got != tc.want[i] {
					t.Errorf("%s offered=%v want=%v", candidates[i].Name, got, tc.want[i])
				}
			}
			for _, choice := range result.choices {
				if slices.Contains(choice, game.PlayerTarget(game.Player3)) {
					t.Fatal("homogeneous permanent union offered a player")
				}
			}
		})
	}
}

func TestCompiledPermanentUnionTargetResolution(t *testing.T) {
	for _, tc := range []struct {
		text  string
		check func(*testing.T, *game.Game, *game.Permanent)
	}{
		{"Return target creature or Vehicle to its owner's hand.", func(t *testing.T, g *game.Game, p *game.Permanent) {
			if !g.Players[game.Player2].Hand.Contains(p.CardInstanceID) || g.Players[game.Player3].Hand.Contains(p.CardInstanceID) {
				t.Fatal("return did not use physical owner")
			}
		}},
		{"Tap target creature or Vehicle.", func(t *testing.T, _ *game.Game, p *game.Permanent) {
			if !p.Tapped {
				t.Fatal("union tap did not resolve")
			}
		}},
		{"Untap target creature or Vehicle.", func(t *testing.T, _ *game.Game, p *game.Permanent) {
			if p.Tapped {
				t.Fatal("union untap did not resolve")
			}
		}},
		{"Exile target creature or Vehicle.", func(t *testing.T, g *game.Game, p *game.Permanent) {
			if !g.Players[game.Player2].Exile.Contains(p.CardInstanceID) {
				t.Fatal("union exile did not resolve to owner's exile")
			}
		}},
		{"Destroy target creature or Vehicle.", func(t *testing.T, g *game.Game, p *game.Permanent) {
			if !g.Players[game.Player2].Graveyard.Contains(p.CardInstanceID) {
				t.Fatal("union destroy did not resolve to owner's graveyard")
			}
		}},
		{"Put two +1/+1 counters on target creature or Vehicle.", func(t *testing.T, _ *game.Game, p *game.Permanent) {
			if p.Counters.Get(counter.PlusOnePlusOne) != 2 {
				t.Fatal("shared union changed old counter placement")
			}
		}},
	} {
		for _, kind := range []string{"Creature", "Vehicle"} {
			t.Run(tc.text+"/"+kind, func(t *testing.T) {
				def := compiledPermanentUnion(t, tc.text)
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				face := game.CardFace{Name: kind, Types: []types.Card{types.Creature}}
				if kind == "Vehicle" {
					face.Types, face.Subtypes = []types.Card{types.Artifact}, []types.Sub{types.Vehicle}
				}
				target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: face})
				target.Controller, target.Tapped = game.Player3, true
				spell := addCardToHand(g, game.Player1, def)
				g.Players[game.Player1].Hand.Remove(spell)
				g.Stack.Push(&game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, SourceCardID: spell, Controller: game.Player1,
					Targets: []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
				})
				NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
				tc.check(t, g, target)
			})
		}
	}
}

func TestCompiledPermanentUnionTargetOwnerControllerAndLifetime(t *testing.T) {
	def := compiledPermanentUnion(t, "Target player gains 1 life. Return target creature or Vehicle to its owner's hand. Its controller loses 2 life. Its owner gains 3 life.")
	for _, lifetime := range []string{"live", "clone", "departed", "reentered", "type changed"} {
		t.Run(lifetime, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addVehiclePermanent(g, game.Player2)
			target.Controller = game.Player3
			decoy := addVehiclePermanent(g, game.Player4)
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(spell)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, SourceCardID: spell, Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player4), game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1, 1},
			}

			g.Stack.Push(obj)
			original := g
			if lifetime == "clone" {
				g = g.Clone()
			}
			if lifetime == "type changed" {
				g.CardInstances[target.CardInstanceID].Def = &game.CardDef{CardFace: game.CardFace{
					Name: "Not a Vehicle", Types: []types.Card{types.Artifact},
				}}
			}
			if lifetime == "departed" || lifetime == "reentered" {
				if !movePermanentToZone(g, permanentForCard(g, target.CardInstanceID), zone.Exile) {
					t.Fatal("target departure failed")
				}
				if lifetime == "reentered" {
					resolver := effectResolver{game: g, obj: obj}
					entered, ok := resolver.putResolvedCardOnBattlefieldValue(g.CardInstances[target.CardInstanceID],
						zone.Exile, game.Player3, nil, permanentCreationOptions{})
					if !ok || entered.ObjectID == target.ObjectID {
						t.Fatal("target reentry did not change incarnation")
					}
				}
			}
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if lifetime == "clone" && permanentForCard(original, target.CardInstanceID) == nil {
				t.Fatal("resolving a clone mutated the original")
			}
			valid := lifetime == "live" || lifetime == "clone"
			if g.Players[game.Player2].Hand.Contains(target.CardInstanceID) != valid {
				t.Fatal("return followed a stale incarnation or missed live owner")
			}
			for player, state := range g.Players {
				want := 40
				if game.PlayerID(player) == game.Player4 {
					want++
				}
				if valid && game.PlayerID(player) == game.Player2 {
					want += 3
				}
				if valid && game.PlayerID(player) == game.Player3 {
					want -= 2
				}
				if state.Life != want {
					t.Errorf("player=%d life=%d want=%d", player, state.Life, want)
				}
			}
			if permanentForCard(g, decoy.CardInstanceID) == nil {
				t.Fatal("union action followed an independent target")
			}
		})
	}
}

func TestCompiledClauseDeclaredTargetAfterCardExile(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		for _, lifetime := range []string{"live", "departed"} {
			t.Run(noun+"/"+lifetime, func(t *testing.T) {
				def := compiledPermanentUnion(t, "Exile target card from a graveyard. Return target "+noun+" to its owner's hand. You gain 1 life.")
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				first := addCardToGraveyard(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Unrelated Card"}})
				target := addCreaturePermanent(g, game.Player2)
				if noun == "creature or Vehicle" {
					target = addVehiclePermanent(g, game.Player2)
				}
				target.Controller = game.Player3
				spell := addCardToHand(g, game.Player1, def)
				g.Players[game.Player1].Hand.Remove(spell)
				g.Stack.Push(&game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, SourceCardID: spell, Controller: game.Player1,
					Targets: []game.Target{currentCardTarget(t, g, first), game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1, 1},
				})
				if lifetime == "departed" && !movePermanentToZone(g, target, zone.Graveyard) {
					t.Fatal("declared target departure failed")
				}
				NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player4].Exile.Contains(first) || g.Players[game.Player4].Hand.Contains(first) {
					t.Fatal("Return stole the unrelated card-result subject")
				}
				if g.Players[game.Player2].Hand.Contains(target.CardInstanceID) != (lifetime == "live") ||
					g.Players[game.Player3].Hand.Contains(target.CardInstanceID) {
					t.Fatal("declared Return lost owner/incarnation identity")
				}
				if g.Players[game.Player1].Life != 41 || g.Players[game.Player2].Life != 40 ||
					g.Players[game.Player3].Life != 40 || g.Players[game.Player4].Life != 40 {
					t.Fatal("independent rider borrowed an object owner/controller")
				}
			})
		}
	}
}
