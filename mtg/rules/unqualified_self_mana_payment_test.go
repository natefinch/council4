package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Canonical #3290; CR 2026-09-25 700.7, 601.2g/h, 605 and 608.2g.
func TestUnqualifiedSelfManaCastAndResolutionPayment(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"cast", "resolution", "resolution after type loss", "resolution clone", "tapped", "changed controller"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := &game.CardDef{CardFace: game.CardFace{
				Name: "Treasure", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Treasure},
				ManaAbilities: []game.ManaAbility{game.TapSacrificeAnyColorManaAbility("{T}, Sacrifice this token: Add one mana of any color.")},
			}}
			tokens, ok := createTokenPermanentsCollectingWithChoices(engine, g, game.Player2, def, 2, false, [game.NumPlayers]PlayerAgent{}, nil)
			if !ok {
				t.Fatal("real token creation failed")
			}
			for _, token := range tokens {
				if scenario == "resolution after type loss" {
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), Layer: game.LayerType, AffectedObjectID: token.ObjectID,
						SetTypes: []types.Card{types.Enchantment}, Duration: game.DurationPermanent,
					})
				}
				if scenario == "changed controller" {
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), Layer: game.LayerControl, AffectedObjectID: token.ObjectID,
						NewController: opt.Val(game.Player3), Duration: game.DurationPermanent,
					})
				}
			}
			if scenario == "tapped" {
				tokens[0].Tapped = true
			}
			spell := &game.CardDef{CardFace: game.CardFace{
				Name: "Colored payment fixture", Types: []types.Card{types.Instant},
				ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
				SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Fixed(2),
				}}}}.Ability()),
			}}
			if scenario == "cast" {
				card := addCardToHand(g, game.Player2, spell)
				setMainPhasePriority(g, game.Player2)
				act := action.CastSpell(card, nil, 0, nil)
				if !containsAction(engine.legalActions(g, game.Player2), act) || !engine.applyAction(g, game.Player2, act) {
					t.Fatal("actual cast payment did not use unqualified mana sources")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
			} else {
				spell.SpellAbility = opt.Val(game.Mode{Sequence: []game.Instruction{
					{Primitive: game.Pay{Payment: game.ResolutionPayment{ManaCost: opt.Val(cost.Mana{cost.U, cost.U})}}, PublishResult: "paid"},
					{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriTrue})},
				}}.Ability())
				addImplementationSpellToStack(g, game.Player2, spell, nil)
				agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}}
				if scenario == "resolution clone" {
					original := g
					g = engine.Simulator().ResolvePriority(g, agents)
					if original.Players[game.Player2].Life != 40 || countTokenPermanentsNamed(original, "Treasure") != 2 {
						t.Fatal("Simulator payment mutated input")
					}
				} else {
					engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				}
			}
			if scenario == "tapped" || scenario == "changed controller" {
				if g.Players[game.Player2].Life != 40 || countTokenPermanentsNamed(g, "Treasure") != 2 {
					t.Fatal("unpayable cost consumed a partial sacrifice/payment")
				}
				return
			}
			if g.Players[game.Player2].Life != 42 || g.Players[game.Player2].ManaPool.Amount(mana.U) != 0 ||
				slices.ContainsFunc(g.Battlefield, func(p *game.Permanent) bool { return p.Token && p.TokenDef.Name == "Treasure" }) {
				t.Fatal("atomic colored cast/resolution payment did not tap, sacrifice and consume both sources")
			}
		})
	}
}
