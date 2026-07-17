package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type electrodominanceAgent struct {
	want string
}

func (electrodominanceAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[0]
}

func (a electrodominanceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	for _, option := range request.Options {
		if strings.Contains(option.Label, a.want) {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

func electrodominanceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Electrodominance",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.X, cost.R, cost.R}),
		Colors:   []color.Color{color.Red},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "any target",
			}},
			Sequence: []game.Instruction{
				{
					Primitive: game.Damage{
						Amount:    game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
						Recipient: game.AnyTargetDamageRecipient(0),
					},
				},
				{
					Optional: true,
					Primitive: game.CastForFree{
						Player: game.ControllerReference(),
						Selection: game.Selection{
							ExcludedTypes: []types.Card{types.Land},
						},
						Zone:              zone.Hand,
						MaxManaValueFromX: true,
					},
				},
			},
		}.Ability()),
	}}
}

func pushElectrodominance(g *game.Game, target game.Target, x int) {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID:    sourceID,
		Owner: game.Player1,
		Def:   electrodominanceDef(),
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     sourceID,
		Controller:   game.Player1,
		Targets:      []game.Target{target},
		TargetCounts: []int{1},
		XValue:       x,
	})
}

func TestElectrodominanceDamageReplacementThenFreeCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Three Drop",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})
	before := g.Players[game.Player2].Life
	pushElectrodominance(g, game.PlayerTarget(game.Player2), 3)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: electrodominanceAgent{want: "Three Drop"}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := before - g.Players[game.Player2].Life; got != 6 {
		t.Fatalf("damage dealt = %d, want 6 after replacement", got)
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != spellID || top.XValue != 0 {
		t.Fatalf("free-cast spell = %+v, want Three Drop with X=0", top)
	}
}

func TestElectrodominanceFizzlesBeforeFreeCastWhenTargetIllegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "One Drop",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}})
	pushElectrodominance(g, game.PermanentTarget(target.ObjectID), 1)
	movePermanentToZone(g, target, zone.Graveyard)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: electrodominanceAgent{want: "One Drop"},
	}, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell left hand after Electrodominance fizzled")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
}

func TestElectrodominanceOptionalFreeCastCanBeDeclined(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addFreeInstant(g, game.Player1, "Declined Spell")
	pushElectrodominance(g, game.PlayerTarget(game.Player2), 0)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: optionalMayAgent{accept: false},
	}, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("declined spell left hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
}

func TestElectrodominanceXZeroCanCastZeroManaSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addFreeInstant(g, game.Player1, "Zero-Mana Spell")
	pushElectrodominance(g, game.PlayerTarget(game.Player2), 0)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: electrodominanceAgent{want: "Zero-Mana Spell"},
	}, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != spellID {
		t.Fatalf("free-cast spell = %+v, want zero-mana spell", top)
	}
}

func TestElectrodominanceCopyPreservesSourceXForFreeCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Copied X Three Drop",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})
	pushElectrodominance(g, game.PlayerTarget(game.Player2), 3)
	copyObject, _ := g.Stack.Peek()
	copyObject.Copy = true

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: electrodominanceAgent{want: "Copied X Three Drop"},
	}, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != spellID || top.XValue != 0 {
		t.Fatalf("copy free-cast spell = %+v, want three-drop with its own X=0", top)
	}
}

func TestXBoundedFreeCastUsesChosenSpellFaceAndFreeCastXZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	selection := game.Selection{ExcludedTypes: []types.Card{types.Land}}
	tests := []struct {
		name     string
		def      *game.CardDef
		wantFace game.FaceIndex
	}{
		{
			name: "modal DFC back",
			def: &game.CardDef{
				CardFace: game.CardFace{Name: "Large Front", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(5)})},
				Layout:   game.LayoutModalDFC,
				Back:     opt.Val(game.CardFace{Name: "Small Back", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.O(2)})}),
			},
			wantFace: game.FaceBack,
		},
		{
			name: "split alternate",
			def: &game.CardDef{
				CardFace: game.CardFace{Name: "Large Half", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(5)})},
				Layout:   game.LayoutSplit,
				Alternate: opt.Val(game.CardFace{
					Name: "Small Half", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.O(2)}),
				}),
			},
			wantFace: game.FaceAlternate,
		},
		{
			name: "adventure alternate",
			def: &game.CardDef{
				CardFace: game.CardFace{Name: "Large Creature", Types: []types.Card{types.Creature}, ManaCost: opt.Val(cost.Mana{cost.O(4)})},
				Layout:   game.LayoutAdventure,
				Alternate: opt.Val(game.CardFace{
					Name: "Small Adventure", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.O(2)}),
				}),
			},
			wantFace: game.FaceAlternate,
		},
		{
			name: "X spell",
			def: &game.CardDef{CardFace: game.CardFace{
				Name: "Variable Spell", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
			}},
			wantFace: game.FaceFront,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cardID := addCardToHand(g, game.Player1, tc.def)
			card, _ := g.GetCardInstance(cardID)
			options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Hand, selection, opt.Val(2))
			if len(options) == 0 {
				t.Fatal("no bounded free-cast option")
			}
			for _, option := range options {
				if option.cast.Face == tc.wantFace && option.cast.XValue == 0 {
					return
				}
			}
			t.Fatalf("options = %#v, want face %v with X=0", options, tc.wantFace)
		})
	}
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})
	land, _ := g.GetCardInstance(landID)
	if options := engine.freeCastOptionsForCard(g, game.Player1, land, zone.Hand, selection, opt.Val(20)); len(options) != 0 {
		t.Fatalf("land options = %#v, want none", options)
	}
}

func TestFreeCastManaValueCapAppliesToChosenFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToExile(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Small Front",
			Types:    []types.Card{types.Sorcery},
			ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:     "Expensive Back",
			Types:    []types.Card{types.Instant},
			ManaCost: opt.Val(cost.Mana{cost.O(7)}),
		}),
	})

	if !engine.castFreeSpellFromExileWithMaxManaValue(g, game.Player1, cardID, 2, [game.NumPlayers]PlayerAgent{
		game.Player1: electrodominanceAgent{want: "Expensive Back"},
	}, &TurnLog{}) {
		t.Fatal("bounded free cast failed")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Face != game.FaceFront {
		t.Fatalf("bounded free cast face = %+v, want eligible front face", top)
	}
}

func TestFreeCastUsesIndependentGraveyardPermission(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Flashback Card",
		Types: []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.R}}},
		}},
	}})

	if !engine.castFreeTargetedSpell(g, game.Player1, cardID, zone.Graveyard, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("effect-granted free cast from graveyard failed")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != cardID || top.Flashback {
		t.Fatalf("free-cast graveyard spell = %+v, want independent non-flashback cast", top)
	}
}

func TestFreeCastPaysAdditionalCostsAndOffersKickerModesTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	discardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Discard Fodder", Types: []types.Card{types.Land},
	}})
	target := addCreaturePermanent(g, game.Player2)
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Modal Kicker Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalDiscard, Amount: 1, Source: zone.Hand,
		}},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1,
			MaxModes: 1,
			Modes: []game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}},
				{
					Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
					Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
				},
			},
		}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}}},
		}},
	}}
	spellID := addCardToHand(g, game.Player1, spell)
	obj := &game.StackObject{Controller: game.Player1, XValue: 2}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: electrodominanceAgent{want: "modes [1] with kicker"}}
	instruction := &game.Instruction{Primitive: game.CastForFree{
		Player: game.ControllerReference(),
		Selection: game.Selection{
			ExcludedTypes: []types.Card{types.Land},
		},
		Zone:              zone.Hand,
		MaxManaValueFromX: true,
	}}

	engine.resolveInstructionWithChoices(g, obj, instruction, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("free-cast spell remained in hand")
	}
	if g.Players[game.Player1].Hand.Contains(discardID) {
		t.Fatal("mandatory discard additional cost was not paid")
	}
	top, ok := g.Stack.Peek()
	if !ok || !top.KickerPaid || len(top.ChosenModes) != 1 || top.ChosenModes[0] != 1 ||
		len(top.Targets) != 1 || top.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("announced free cast = %+v", top)
	}
}

func TestFreeCastChoosesPayableAdditionalCostBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 3
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Additional Choice Spell",
		Types: []types.Card{types.Instant},
		AdditionalCostChoices: []cost.AdditionalChoice{{
			Options: []cost.AdditionalChoiceOption{
				{
					Label: "Pay 5 life",
					Costs: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 5}},
				},
				{Label: "Pay {2}", Mana: cost.Mana{cost.O(2)}},
			},
		}},
	}})
	obj := &game.StackObject{Controller: game.Player1, XValue: 0}
	instruction := &game.Instruction{Primitive: game.CastForFree{
		Player:            game.ControllerReference(),
		Selection:         game.Selection{ExcludedTypes: []types.Card{types.Land}},
		Zone:              zone.Hand,
		MaxManaValueFromX: true,
	}}

	engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{
		game.Player1: electrodominanceAgent{want: "Additional Choice Spell"},
	}, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != spellID {
		t.Fatalf("free-cast spell = %+v, want additional-choice spell", top)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool = %d, want pay-mana branch consumed 2", got)
	}
	if got := g.Players[game.Player1].Life; got != 3 {
		t.Fatalf("life = %d, want unaffordable life branch skipped", got)
	}
}

func TestBoundedFreeCastHonorsRestrictionsProhibitionsAndLimits(t *testing.T) {
	t.Run("normal timing ignored", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Turn.ActivePlayer = game.Player2
		g.Turn.Phase = game.PhaseCombat
		cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Off-Turn Sorcery",
			Types: []types.Card{types.Sorcery},
		}})
		card, _ := g.GetCardInstance(cardID)
		if options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Hand, game.Selection{}, opt.Val(0)); len(options) == 0 {
			t.Fatal("normal sorcery timing incorrectly blocked free cast")
		}
	})
	t.Run("printed restriction", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Restricted Spell",
			Types: []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{{
				CastOnlyAfterAttackedThisStep: true,
			}},
		}})
		card, _ := g.GetCardInstance(cardID)
		if options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Hand, game.Selection{}, opt.Val(0)); len(options) != 0 {
			t.Fatalf("restricted options = %#v", options)
		}
	})
	t.Run("prohibition", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addFreeInstant(g, game.Player1, "Prohibited Spell")
		card, _ := g.GetCardInstance(cardID)
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCantCastSpells,
			Controller:     game.Player1,
			AffectedPlayer: game.PlayerYou,
		})
		if options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Hand, game.Selection{}, opt.Val(0)); len(options) != 0 {
			t.Fatalf("prohibited options = %#v", options)
		}
	})
	t.Run("per-turn limit", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		castLimitPermanent(g, game.Player1, game.PlayerYou, 1)
		castSpellTargeting(g, game.Player1)
		cardID := addFreeInstant(g, game.Player1, "Limited Spell")
		card, _ := g.GetCardInstance(cardID)
		if options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Hand, game.Selection{}, opt.Val(0)); len(options) != 0 {
			t.Fatalf("limited options = %#v", options)
		}
	})
}
