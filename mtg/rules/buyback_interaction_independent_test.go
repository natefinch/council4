package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Independent review against the pinned Oracle printings and full saved rulings
// for Reiterate, Allay, and Capsize; CR effective 2026-09-25 (verified SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Relevant reads: 118.3/8, 304.1, 400.3/7, 601.2, 608.2, 614.1/4/15,
// 616.1, 701.8/19, 702.11/12/16/27a, 707.10, and 903.9.

func independentInteractionDef(t *testing.T, registry *cards.Registry, oracleID string) *game.CardDef {
	t.Helper()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil || def == nil {
		t.Fatalf("checked production identity %s: %v", oracleID, err)
	}
	return def
}

func TestBuybackInteractionIndependentSourcePaymentAndUntargetedSorcery(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, oracleID string
		land           types.Sub
		base           int
	}{
		{"Reiterate", "38394ee3-6726-4f91-bc25-36bce0c6aab9", types.Mountain, 3},
		{"Allay", "664d982a-5cd8-4289-afb8-fa2e083042c9", types.Plains, 2},
		{"Capsize", "77637eff-2963-4402-88f3-ca346f762fc8", types.Island, 3},
	} {
		for _, paid := range []bool{false, true} {
			name := tc.name + "/normal"
			if paid {
				name = tc.name + "/buyback"
			}
			t.Run(name, func(t *testing.T) {
				registry := cards.NewDefaultRegistry()
				engine := NewEngineWithPrograms(nil, registry.Programs())
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := game.Target{}
				var targetPermanent *game.Permanent
				setMainPhasePriority(g, game.Player2)
				if tc.name == "Reiterate" {
					sorcery := &game.CardDef{CardFace: game.CardFace{
						Name: "Independent untargeted sorcery", Types: []types.Card{types.Sorcery},
						ManaCost: opt.Val(cost.Mana{cost.O(0)}),
						SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
							Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(7)},
						}}}.Ability()),
					}}
					subjectID := addCardToHand(g, game.Player2, sorcery)
					if !engine.applyAction(g, game.Player2, action.CastSpell(subjectID, nil, 0, nil)) {
						t.Fatal("otherwise legal untargeted sorcery failed to cast")
					}
					subject, _ := g.Stack.Peek()
					target = game.StackObjectTarget(subject.ID)
				} else {
					targetPermanent = addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
						Name: "Independent enchantment", Types: []types.Card{types.Enchantment},
					}})
					target = game.PermanentTarget(targetPermanent.ObjectID)
				}
				for range tc.base + 3 {
					addBasicLandPermanent(g, game.Player1, tc.land)
				}
				cardID := addCardToHand(g, game.Player1, independentInteractionDef(t, registry, tc.oracleID))
				g.Turn.PriorityPlayer = game.Player1
				act := action.CastSpell(cardID, []game.Target{target}, 0, nil)
				offers := 0
				agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
					if request.Kind == game.ChoicePayment && request.Prompt == "Choose spell cost" {
						offers++
						if len(request.Options) != 2 || g.Players[game.Player1].ManaPool.Total() != 0 {
							t.Fatal("normal and buyback must be offered before tapping mana sources")
						}
						for _, option := range request.Options {
							if strings.Contains(option.Label, "Buyback") == paid {
								return []int{option.Index}
							}
						}
						t.Fatal("requested additional-cost choice missing")
					}
					return request.DefaultSelection
				}}}
				before := g.Clone()
				if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
					t.Fatal("registered spell could not defer mana production until payment")
				}
				spell, _ := g.Stack.Peek()
				wantSpent := tc.base
				if paid {
					wantSpent += 3
				}
				tapped := 0
				for _, permanent := range g.Battlefield {
					if permanent.Controller == game.Player1 && permanent.Tapped {
						tapped++
					}
				}
				if offers != 1 || spell.BuybackPaid != paid || spell.ManaSpentToCast != wantSpent || tapped != wantSpent {
					t.Fatalf("deferred payment: offers=%d paid=%v spent=%d tapped=%d; want paid=%v spent=%d",
						offers, spell.BuybackPaid, spell.ManaSpentToCast, tapped, paid, wantSpent)
				}
				simulated, accepted := engine.Simulator().Apply(before, game.Player1, act,
					buybackInteractionChoiceAgent(t, game.Player1, paid))
				if !accepted || simulated == nil {
					t.Fatal("Simulator rejected the payable source-funded cast")
				}
				for g.Stack.Size() != 0 {
					engine.resolveTopOfStack(g, &TurnLog{})
				}
				simulated = engine.Simulator().ResolvePriority(simulated, simPassPolicies())
				for name, result := range map[string]*game.Game{"live": g, "Simulator": simulated} {
					if result.Players[game.Player1].Hand.Contains(cardID) != paid ||
						result.Players[game.Player1].Graveyard.Contains(cardID) == paid {
						t.Fatalf("%s: wrong normal/buyback destination", name)
					}
					switch tc.name {
					case "Reiterate":
						if result.Players[game.Player1].Life != 47 || result.Players[game.Player2].Life != 47 ||
							countEventsOfKind(result.Events, game.EventSpellCast) != 2 ||
							countEventsOfKind(result.Events, game.EventSpellCopied) != 1 {
							t.Fatalf("%s: untargeted sorcery was not copied under the copier's control", name)
						}
					case "Allay":
						if !result.Players[game.Player2].Graveyard.Contains(targetPermanent.CardInstanceID) {
							t.Fatalf("%s: enchantment was not destroyed", name)
						}
					case "Capsize":
						if !result.Players[game.Player2].Hand.Contains(targetPermanent.CardInstanceID) {
							t.Fatalf("%s: permanent did not reach its owner's hand", name)
						}
					}
				}
				if !before.Players[game.Player1].Hand.Contains(cardID) {
					t.Fatal("Simulator mutated the source position")
				}
			})
		}
	}
}

func TestBuybackInteractionIndependentReiterateKeepsIllegalCopiedTarget(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Controller-relative hexproof", Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{game.HexproofStaticBody},
	}})
	capsizeID, capsize := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
		game.Player2, game.PermanentTarget(target.ObjectID), true,
		[]mana.Color{mana.U, mana.U, mana.C, mana.C, mana.C, mana.C})
	reiterateID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[0],
		game.Player1, game.StackObjectTarget(capsize.ID), true,
		[]mana.Color{mana.R, mana.R, mana.C, mana.C, mana.C, mana.C})
	engine.resolveTopOfStack(g, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.Controller != game.Player1 || !copied.BuybackPaid ||
		len(copied.Targets) != 1 || copied.Targets[0].PermanentID != target.ObjectID {
		t.Fatal("Reiterate must create the copy and may retain its now-illegal inherited target")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, live := permanentByObjectID(g, target.ObjectID); !live ||
		g.Players[game.Player2].Hand.Contains(capsizeID) ||
		!g.Players[game.Player1].Hand.Contains(reiterateID) {
		t.Fatal("all-illegal copy affected the target or moved its original physical card")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
		!g.Players[game.Player2].Hand.Contains(capsizeID) ||
		countEventsOfKind(g.Events, game.EventSpellCast) != 2 ||
		countEventsOfKind(g.Events, game.EventSpellCopied) != 1 {
		t.Fatal("original positive control must still resolve and buy back")
	}
}

func TestBuybackInteractionIndependentAllayResolutionRevalidation(t *testing.T) {
	t.Parallel()
	for _, change := range []string{"loses enchantment", "gains protection", "gains indestructible"} {
		t.Run(change, func(t *testing.T) {
			registry, engine := buybackInteractionRuntime(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Initially legal enchantment", Types: []types.Card{types.Enchantment},
			}})
			cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[1],
				game.Player1, game.PermanentTarget(target.ObjectID), true,
				[]mana.Color{mana.W, mana.C, mana.C, mana.C, mana.C})
			control := g.Clone()
			engine.resolveTopOfStack(control, &TurnLog{})
			if !control.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
				!control.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("otherwise identical positive control failed")
			}
			effect := game.ContinuousEffect{Layer: game.LayerAbility}
			switch change {
			case "loses enchantment":
				effect.Layer, effect.SetTypes = game.LayerType, []types.Card{types.Artifact}
			case "gains protection":
				effect.AddAbilities = []game.Ability{&game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{game.ProtectionKeyword{FromColors: []color.Color{color.White}}},
				}}
			case "gains indestructible":
				effect.AddKeywords = []game.Keyword{game.Indestructible}
			}
			addEffectSpellToStack(g, game.Player2, game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)), ContinuousEffects: []game.ContinuousEffect{effect},
				Duration: game.DurationPermanent,
			}, []game.Target{game.PermanentTarget(target.ObjectID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.resolveTopOfStack(g, &TurnLog{})
			wantReturn := change == "gains indestructible"
			if _, live := permanentByObjectID(g, target.ObjectID); !live ||
				g.Players[game.Player1].Hand.Contains(cardID) != wantReturn ||
				g.Players[game.Player1].Graveyard.Contains(cardID) == wantReturn {
				t.Fatal("legal-but-ineffective destruction and all-illegal resolution must differ")
			}
		})
	}
}

func TestBuybackInteractionIndependentCapsizeCommanderOwnerDeclinesAfterControlChange(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCommanderPermanent(g, game.Player2)
	cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
		game.Player1, game.PermanentTarget(target.ObjectID), true,
		[]mana.Color{mana.U, mana.U, mana.C, mana.C, mana.C, mana.C})
	addEffectSpellToStack(g, game.Player3, game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerControl, NewController: opt.Val(game.Player3),
		}},
		Duration: game.DurationPermanent,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if effectiveController(g, target) != game.Player3 {
		t.Fatal("response failed to change the target's controller")
	}
	offers := 0
	owner := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceMay {
			offers++
			if request.Player != game.Player2 {
				t.Fatal("commander choice belongs to its owner, not its current controller")
			}
			return []int{0}
		}
		return request.DefaultSelection
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: owner}, &TurnLog{})
	if offers != 1 || !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
		g.Players[game.Player3].Hand.Contains(target.CardInstanceID) ||
		g.Players[game.Player2].CommandZone.Contains(target.CardInstanceID) ||
		!g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("Capsize must honor the owner's decline and buy back independently")
	}
}

func TestBuybackInteractionIndependentCapsizeTokenPriorityCleanup(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addEffectSpellToStack(g, game.Player2, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenDef(&game.CardDef{CardFace: game.CardFace{
			Name: "Independent artifact token", Types: []types.Card{types.Artifact},
		}}),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != 1 || !g.Battlefield[0].Token {
		t.Fatal("Engine token creation failed")
	}
	token := g.Battlefield[0]
	cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
		game.Player1, game.PermanentTarget(token.ObjectID), true,
		[]mana.Color{mana.U, mana.U, mana.C, mana.C, mana.C, mana.C})
	result := engine.Simulator().ResolvePriority(g, simPassPolicies())
	reachedHand := false
	for _, event := range result.Events {
		if event.Kind == game.EventZoneChanged && event.PermanentID == token.ObjectID && event.ToZone == zone.Hand {
			reachedHand = true
		}
	}
	if _, live := permanentByObjectID(result, token.ObjectID); live || !reachedHand ||
		result.Players[game.Player2].Hand.Contains(token.ObjectID) ||
		!result.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("returned token must reach hand, then cease to exist before priority, while Capsize buys back")
	}
	if _, live := permanentByObjectID(g, token.ObjectID); !live {
		t.Fatal("Simulator cleanup mutated its source game")
	}
}
