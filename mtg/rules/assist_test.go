package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func assistTestSpell() *game.CardDef {
	def := simpleGainLifeInstant("Assist test spell")
	def.ManaCost = opt.Val(cost.Mana{cost.O(3), cost.W})
	def.StaticAbilities = []game.StaticAbility{game.AssistStaticBody}
	return def
}

func TestAssistManaWindowsAndReplay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := assistTestSpell()
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
		Sequence: []game.Instruction{{Primitive: game.Destroy{
			Object: game.TargetPermanentReference(0),
		}}},
	}.Ability())
	target := targetsSourceTaxPermanent(g, game.Player3, 1)
	targetCard, _ := g.GetCardInstance(target.CardInstanceID)
	targetCard.Def.Types = []types.Card{types.Artifact}
	_ = addNamedTypedPermanent(g, game.Player4, "Unchosen target", types.Artifact)
	white := addBasicLandPermanent(g, game.Player1, types.Plains)
	blue := addBasicLandPermanent(g, game.Player2, types.Island)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	spell := addCardToHand(g, game.Player1, def)
	act := action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	cloned := g.Clone()
	type decision struct {
		request game.ChoiceRequest
		answer  []int
	}
	var decisions []decision
	play := func(g *game.Game, replay bool) *TurnLog {
		t.Helper()
		next := 0
		choose := func(req game.ChoiceRequest) []int {
			if replay {
				if next >= len(decisions) || !reflect.DeepEqual(req, decisions[next].request) {
					t.Fatalf("replay choice %d differs: %+v", next, req)
				}
				next++
				return decisions[next-1].answer
			}
			answer := req.DefaultSelection
			switch req.Prompt {
			case assistPlayerPrompt, assistManaPrompt, assistCasterManaPrompt, assistContributionPrompt:
				obj, _ := g.Stack.Peek()
				if obj == nil || !obj.Casting || !slices.Equal(obj.Targets, []game.Target{game.PermanentTarget(target.ObjectID)}) || spellCastEventCount(g) != 0 {
					t.Fatalf("Assist prompt %q ran outside the announced cast", req.Prompt)
				}
			}
			switch req.Prompt {
			case assistPlayerPrompt:
				if req.Player != game.Player1 || len(g.ManaPaymentScopes) != 0 {
					t.Fatal("caster must invite the helper before either window")
				}
			case assistManaPrompt:
				w, _ := permanentByObjectID(g, white.ObjectID)
				if req.Player != game.Player2 || !req.Private || !g.ManaPaymentAllowed(game.Player2) || w.Tapped {
					t.Fatal("helper's private mana window must precede the caster's")
				}
			case assistCasterManaPrompt:
				b, _ := permanentByObjectID(g, blue.ObjectID)
				if req.Player != game.Player1 || !req.Private || !g.ManaPaymentAllowed(game.Player1) || !b.Tapped {
					t.Fatal("the caster's mana window must follow the helper's")
				}
			case assistContributionPrompt:
				w, _ := permanentByObjectID(g, white.ObjectID)
				b, _ := permanentByObjectID(g, blue.ObjectID)
				if req.Player != game.Player2 || !w.Tapped || !b.Tapped || len(g.ManaPaymentScopes) != 0 ||
					req.NumberRange.Val.Maximum.Val != 4 || g.Players[game.Player1].ManaPool.Amount(mana.W) != 1 {
					t.Fatalf("contribution out of sequence: white=%t blue=%t scopes=%v max=%d mana=%d",
						w.Tapped, b.Tapped, g.ManaPaymentScopes, req.NumberRange.Val.Maximum.Val, g.Players[game.Player1].ManaPool.Amount(mana.W))
				}
			}
			decisions = append(decisions, decision{req, answer})
			return answer
		}
		var agents [game.NumPlayers]PlayerAgent
		for _, seat := range []game.PlayerID{game.Player1, game.Player2} {
			agents[seat] = &authoredCopyChoiceAgent{choose: choose}
		}
		log := &TurnLog{}
		if !slices.ContainsFunc(e.legalActions(g, game.Player1), func(candidate action.Action) bool {
			cast, ok := candidate.CastSpellPayload()
			return ok && cast.CardID == spell && slices.Equal(cast.Targets, []game.Target{game.PermanentTarget(target.ObjectID)})
		}) {
			t.Fatal("the legally assistable cast was not enumerated")
		}
		if !e.applyActionWithChoices(g, game.Player1, act, agents, log) {
			t.Fatal("assist cast with tapped mana sources rejected")
		}
		if replay && next != len(decisions) {
			t.Fatal("replay did not consume all decisions")
		}
		obj, _ := g.Stack.Peek()
		if obj.ManaSpentToCast != 5 || obj.ManaSpentByColorToCast[color.White] != 1 || obj.ManaSpentByColorToCast[color.Blue] != 1 ||
			spellCastEventCount(g) != 1 || g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != 0 {
			t.Fatalf("wrong allocation or cast bookkeeping: %+v", obj)
		}
		if slices.ContainsFunc(log.Choices, func(d game.ChoiceDecision) bool { return d.Request.Prompt == assistManaPrompt }) {
			t.Fatal("private activation menu leaked into the public log")
		}
		e.resolveTopOfStack(g, log)
		if !g.Players[game.Player3].Graveyard.Contains(target.CardInstanceID) {
			t.Fatal("assisted spell did not resolve")
		}
		return log
	}
	first, second := play(g, false), play(cloned, true)
	if !reflect.DeepEqual(first.Choices, second.Choices) {
		t.Fatal("public Assist decisions diverged on replay")
	}
}

func TestAssistCombinedCostsAndDecline(t *testing.T) {
	for _, tc := range []struct {
		name        string
		casterMana  int
		contribute  int
		decline     bool
		wantSuccess bool
	}{
		{"partial contribution", 3, 2, false, true},
		{"helper declines", 5, 0, false, true},
		{"caster declines", 5, 0, true, true},
		{"insufficient offer reverses", 1, 1, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			def := assistTestSpell()
			def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			def.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
				{Label: "Two extra mana", Mana: cost.Mana{cost.O(2)}},
				{Label: "Two life", Costs: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}}},
			}}}
			spell := addCardToHand(g, game.Player1, def)
			discard := addCardToHand(g, game.Player1, vanillaCreature("Discard me", 1, 1))
			setMainPhasePriority(g, game.Player1)
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.casterMana)
			land := addBasicLandPermanent(g, game.Player2, types.Forest)
			g.Players[game.Player2].ManaPool.Add(mana.U, 4)
			beforeID := g.IDGen.Current()
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
				switch req.Prompt {
				case assistPlayerPrompt:
					if tc.decline {
						return []int{0}
					}
					return []int{1}
				case "Choose spell cost":
					for _, option := range req.Options {
						if option.ManaCost.Exists && option.ManaCost.Val.ManaValue() == 6 {
							return []int{option.Index}
						}
					}
					t.Fatal("jointly payable mana choice missing")
				}
				return req.DefaultSelection
			}}
			agents[game.Player2] = &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
				switch req.Prompt {
				case assistManaPrompt:
					return []int{1}
				case assistContributionPrompt:
					if req.NumberRange.Val.Maximum.Val != 5 {
						t.Fatal("additional generic mana missing from the helper's payment limit")
					}
					return []int{tc.contribute}
				}
				return req.DefaultSelection
			}}
			got := e.applyActionWithChoices(g, game.Player1, action.CastSpell(spell, nil, 0, nil), agents, &TurnLog{})
			if got != tc.wantSuccess {
				t.Fatalf("cast applied=%v want %v", got, tc.wantSuccess)
			}
			if !got {
				if !g.Players[game.Player1].Hand.Contains(spell) || !g.Players[game.Player1].Hand.Contains(discard) ||
					land.Tapped || g.Players[game.Player1].ManaPool.Total() != 2 || g.Players[game.Player2].ManaPool.Total() != 4 ||
					g.Stack.Size() != 0 || len(g.Events) != 0 || g.IDGen.Current() != beforeID || len(g.ManaPaymentScopes) != 0 {
					t.Fatal("unsuccessful Assist left a partial payment or mana activation")
				}
				return
			}
			wantHelper := 5 - tc.contribute
			if tc.decline {
				wantHelper = 4
			}
			if !g.Players[game.Player1].Graveyard.Contains(discard) || g.Players[game.Player1].Life != 40 ||
				g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != wantHelper {
				t.Fatal("Assist did not pay the chosen mixed cost or spent declined mana")
			}
		})
	}
}

func TestAssistOnlyOnePlayerMayPayOnlyGeneric(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spell := addCardToHand(g, game.Player1, assistTestSpell())
	act := action.CastSpell(spell, nil, 0, nil)
	g.Players[game.Player2].ManaPool.Add(mana.W, 3)
	g.Players[game.Player3].ManaPool.Add(mana.C, 2)
	g.Players[game.Player4].Eliminated = true
	if e.applyAction(g, game.Player1, act) {
		t.Fatal("Assist paid a white symbol from another player's pool")
	}
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	var agents [game.NumPlayers]PlayerAgent
	invitations := 0
	agents[game.Player1] = &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
		if req.Prompt != assistPlayerPrompt {
			return req.DefaultSelection
		}
		invitations++
		if len(req.Options) != 3 || !slices.Equal(req.Options[2].Targets, []game.Target{game.PlayerTarget(game.Player3)}) {
			t.Fatalf("wrong helper candidates: %+v", req.Options)
		}
		return []int{2}
	}}
	if e.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) || invitations != 1 ||
		g.Players[game.Player2].ManaPool.Amount(mana.W) != 3 || g.Players[game.Player3].ManaPool.Amount(mana.C) != 2 ||
		g.Players[game.Player1].ManaPool.Amount(mana.W) != 1 || !g.Players[game.Player1].Hand.Contains(spell) {
		t.Fatal("Assist combined players' mana, retried with a different helper, or partially paid the rejected cast")
	}
}

func TestAssistForecastsPublicMana(t *testing.T) {
	for _, name := range []string{"Sol Ring", "Mana Vault", "restricted mana", "tapped source", "hidden source", "hidden activation cost"} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			spell := addCardToHand(g, game.Player1, assistTestSpell())
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			helper := game.Player2
			var source, island *game.Permanent
			switch name {
			case "Sol Ring", "Mana Vault":
				source = addCombatPermanent(g, helper, registryCardDef(t, name))
				if name == "Sol Ring" {
					island = addBasicLandPermanent(g, helper, types.Island)
				}
			case "tapped source":
				source = addCombatPermanent(g, helper, registryCardDef(t, "Mana Vault"))
				source.Tapped = true
			case "hidden source":
				g.Players[helper].ManaPool.Add(mana.C, 2)
				addCardToHand(g, helper, spiritGuideHandCard("Hidden source", mana.R))
			default:
				body := game.TapManaAbility(mana.C)
				add := game.AddMana{ManaColor: mana.C, Amount: game.Fixed(3)}
				if name == "restricted mana" {
					add.SpendRider = opt.Val(game.ManaSpendRider{
						Condition: game.ManaSpendCastArtifactSpellOnly, Restriction: game.ManaSpendRestrictedToCondition,
					})
				} else {
					body.AdditionalCosts = append(slices.Clone(body.AdditionalCosts), cost.Additional{Kind: cost.AdditionalDiscard, Amount: 1})
					addCardToHand(g, helper, vanillaCreature("Hidden discard", 1, 1))
				}
				body.Content.Modes[0].Sequence[0].Primitive = add
				source = addCombatPermanent(g, helper, &game.CardDef{CardFace: game.CardFace{
					Name: name, Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{body},
				}})
			}
			before := g.Clone()
			want := name == "Sol Ring" || name == "Mana Vault"
			listed := slices.ContainsFunc(e.legalActions(g, game.Player1), func(a action.Action) bool {
				cast, ok := a.CastSpellPayload()
				return ok && cast.CardID == spell
			})
			if listed != want || !reflect.DeepEqual(before, g.Clone()) {
				t.Fatalf("legal action listing=%v, want %v; the forecast must not change the game", listed, want)
			}
			log := &TurnLog{}
			got := e.applyActionWithChoices(g, game.Player1, action.CastSpell(spell, nil, 0, nil), [game.NumPlayers]PlayerAgent{}, log)
			if got != want {
				t.Fatalf("cast applied=%v, want %v", got, want)
			}
			if !want {
				if !g.Players[game.Player1].Hand.Contains(spell) || g.Players[game.Player1].ManaPool.Amount(mana.W) != 1 {
					t.Fatal("unaffordable cast spent the caster's mana or moved the spell")
				}
				return
			}
			if !source.Tapped || island != nil && !island.Tapped ||
				g.Players[helper].ManaPool.Total() != 0 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("the helper's public activation witness did not pay the generic cost")
			}
			e.resolveTopOfStack(g, log)
			if !g.Players[game.Player1].Graveyard.Contains(spell) {
				t.Fatal("assisted spell did not resolve")
			}
		})
	}
}

func TestAssistPublicManaWitnessUsesNeededColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spell := addCardToHand(g, game.Player1, assistTestSpell())
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	helper := game.Player2
	// The first color cannot pay the next source's activation cost. Both
	// activations are needed, so forecasting the right source alone is not enough.
	choice := game.TapManaChoiceAbility(mana.U, mana.G)
	first := addCombatPermanent(g, helper, &game.CardDef{CardFace: game.CardFace{
		Name: "Two-color source", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{choice},
	}})
	filter := game.TapManaAbility(mana.C)
	filter.ManaCost = opt.Val(cost.Mana{cost.G})
	filter.Content.Modes[0].Sequence[0].Primitive = game.AddMana{ManaColor: mana.C, Amount: game.Fixed(3)}
	second := addCombatPermanent(g, helper, &game.CardDef{CardFace: game.CardFace{
		Name: "Green filter", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{filter},
	}})
	if !e.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Assist did not follow the public color and input-cost witness")
	}
	if !first.Tapped || !second.Tapped || g.Players[helper].ManaPool.Total() != 0 {
		t.Fatal("the helper should tap both sources and spend all three produced mana")
	}
}

func TestAssistCasterFollowsManaActivationWitness(t *testing.T) {
	for _, mode := range []string{"no choice agent", "invalid source choice", "explicit output choice"} {
		t.Run(mode, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			caster, helper := game.Player1, game.Player2
			spell := addCardToHand(g, caster, assistTestSpell())
			vault := addCombatPermanent(g, helper, registryCardDef(t, "Mana Vault"))
			// The caster must choose the second output to fund a second
			// activation, rather than merely choosing the spell's color.
			source := addCombatPermanent(g, caster, &game.CardDef{CardFace: game.CardFace{
				Name: "Caster's two-color source", Types: []types.Card{types.Artifact},
				ManaAbilities: []game.ManaAbility{game.TapManaChoiceAbility(mana.U, mana.G)},
			}})
			filter := game.TapManaAbility(mana.W)
			filter.ManaCost = opt.Val(cost.Mana{cost.G})
			filterSource := addCombatPermanent(g, caster, &game.CardDef{CardFace: game.CardFace{
				Name: "Caster's green filter", Types: []types.Card{types.Artifact},
				ManaAbilities: []game.ManaAbility{filter},
			}})
			if !slices.ContainsFunc(e.legalActions(g, caster), func(candidate action.Action) bool {
				cast, ok := candidate.CastSpellPayload()
				return ok && cast.CardID == spell
			}) {
				t.Fatal("caster's activation chain did not make the assisted spell legal")
			}
			var agents [game.NumPlayers]PlayerAgent
			sourceChoices, outputChoices := 0, 0
			if mode != "no choice agent" {
				agents[caster] = &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
					switch req.Prompt {
					case assistCasterManaPrompt:
						sourceChoices++
						if mode == "invalid source choice" {
							return []int{len(req.Options)}
						}
					case "Choose a color":
						outputChoices++
						// A valid explicit decision can depart from the witness.
						return []int{0}
					}
					return req.DefaultSelection
				}}
			}
			applied := e.applyActionWithChoices(g, caster, action.CastSpell(spell, nil, 0, nil), agents, &TurnLog{})
			if mode == "explicit output choice" {
				if applied || sourceChoices == 0 || outputChoices != 1 || !g.Players[caster].Hand.Contains(spell) ||
					g.Stack.Size() != 0 || g.Players[caster].ManaPool.Total() != 0 || g.Players[helper].ManaPool.Total() != 0 {
					t.Fatal("explicit blue output was ignored or its unsuccessful cast did not roll back")
				}
				return
			}
			if !applied || !source.Tapped || !filterSource.Tapped || !vault.Tapped ||
				g.Players[caster].ManaPool.Total() != 0 || g.Players[helper].ManaPool.Total() != 0 {
				t.Fatal("fallback did not use the caster's witnessed output and activation cost to pay the spell")
			}
			if mode == "invalid source choice" && (sourceChoices != 2 || outputChoices != 0) {
				t.Fatalf("expected two fallback activations with no independent output decision, got %d source and %d output choices", sourceChoices, outputChoices)
			}
		})
	}
}

func TestAssistHandManaOptionIsPrivate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spell := addCardToHand(g, game.Player1, assistTestSpell())
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	unused := addCardToHand(g, game.Player2, spiritGuideHandCard("Unused helper hand source", mana.G))
	used := addCardToHand(g, game.Player2, spiritGuideHandCard("Used helper hand source", mana.R))
	var agents [game.NumPlayers]PlayerAgent
	menuSeen := false
	agents[game.Player2] = &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
		if req.Prompt != assistManaPrompt {
			return req.DefaultSelection
		}
		if !req.Private || req.Player != game.Player2 {
			t.Fatal("helper's hand source options must be private")
		}
		if menuSeen {
			return []int{0}
		}
		menuSeen = true
		for _, option := range req.Options {
			if strings.Contains(option.Label, "Used helper hand source") {
				return []int{option.Index}
			}
		}
		t.Fatal("the invited helper could not activate their hand mana ability")
		return nil
	}}
	log := &TurnLog{}
	if !e.applyActionWithChoices(g, game.Player1, action.CastSpell(spell, nil, 0, nil), agents, log) ||
		!menuSeen || !g.Players[game.Player2].Exile.Contains(used) || !g.Players[game.Player2].Hand.Contains(unused) ||
		g.Players[game.Player2].ManaPool.Total() != 1 {
		t.Fatal("assisted cast did not use the chosen mana ability and pay exactly three generic")
	}
	for _, choice := range log.Choices {
		for _, option := range choice.Request.Options {
			if strings.Contains(option.Label, "Unused helper hand source") {
				t.Fatal("public choices exposed the unactivated hand source")
			}
		}
	}
}

func TestAssistUsesHelperManaSpendRiders(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spell := assistTestSpell()
	sorcery := addCardToHand(g, game.Player1, spell)
	creature := vanillaCreature("Assist creature", 2, 2)
	creature.ManaCost, creature.StaticAbilities = spell.ManaCost, spell.StaticAbilities
	creatureID := addCardToHand(g, game.Player1, creature)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	helper := g.Players[game.Player2]
	helper.ManaPool.Add(mana.C, 2)
	helper.ManaPool.Add(mana.G, 1)
	helper.ManaRiders = []game.ManaRiderInstance{{
		Unit: mana.Unit{Color: mana.G}, Controller: game.Player2,
		Rider: game.ManaSpendRider{
			Condition: game.ManaSpendCastCreatureSpell, Restriction: game.ManaSpendRestrictedToCondition,
			SpellGainsKeywords: []game.Keyword{game.Haste},
		},
	}}
	if e.applyAction(g, game.Player1, action.CastSpell(sorcery, nil, 0, nil)) || helper.ManaPool.Total() != 3 || len(helper.ManaRiders) != 1 {
		t.Fatal("Assist let creature-restricted mana pay for a noncreature spell")
	}
	if !e.applyAction(g, game.Player1, action.CastSpell(creatureID, nil, 0, nil)) {
		t.Fatal("Assist rejected eligible tagged mana for the creature spell")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, creatureID)
	if permanent == nil || !hasKeyword(g, permanent, game.Haste) ||
		helper.ManaPool.Total() != 0 || len(helper.ManaRiders) != 0 {
		t.Fatal("helper's spent mana did not grant haste or its rider survived payment")
	}
}

func TestAssistCastAndCopyRoutes(t *testing.T) {
	for _, route := range []string{"paid exile", "free with surcharge", "free without surcharge", "cast card copy", "prepared copy", "stack copies and storm"} {
		t.Run(route, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			def := assistTestSpell()
			vault := addCombatPermanent(g, game.Player2, registryCardDef(t, "Mana Vault"))
			plateau := addCombatPermanent(g, game.Player1, registryCardDef(t, "Plateau"))
			p := g.Players[game.Player1]
			var cardID id.ID
			var applied bool
			var agents [game.NumPlayers]PlayerAgent
			log := &TurnLog{}
			wantSpent := 4
			if route == "free without surcharge" {
				agents[game.Player1] = &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
					if req.Prompt == assistPlayerPrompt {
						t.Fatal("free spell with no generic cost asked for assistance")
					}
					return req.DefaultSelection
				}}
			}
			switch route {
			case "paid exile":
				cardID = addCardToExile(g, game.Player1, def)
				applied = e.castPaidSpellFromSource(g, p, game.Player1, cardID, zone.Exile, agents, log)
			case "free with surcharge", "free without surcharge", "cast card copy":
				if route == "free without surcharge" {
					wantSpent = 0
				} else {
					def.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
						{Label: "Mana surcharge", Mana: cost.Mana{cost.O(3), cost.W}},
					}}}
				}
				cardID = addCardToExile(g, game.Player1, def)
				if route == "cast card copy" {
					def.StaticAbilities = append(def.StaticAbilities, game.CascadeStaticBody)
					applied = e.castFreeCopyOfCard(g, game.Player1, cardID, agents, log)
				} else {
					card, _ := g.GetCardInstance(cardID)
					applied = e.castFreeOption(g, p, game.Player1, card, action.CastSpellAction{CardID: cardID, SourceZone: zone.Exile}, false, agents, log)
				}
			case "prepared copy":
				prepared := prepareCard()
				prepared.Alternate.Val.StaticAbilities = def.StaticAbilities
				prepared.Alternate.Val.ManaCost = def.ManaCost
				permanent := addCombatPermanent(g, game.Player1, prepared)
				permanent.Prepared = true
				cardID = permanent.CardInstanceID
				applied = e.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil))
				if permanent.Prepared {
					t.Fatal("assisted cast did not consume the prepared state")
				}
			case "stack copies and storm":
				first := simpleGainLifeInstant("Earlier spell")
				first.ManaCost = opt.Val(cost.Mana{})
				earlier := addCardToHand(g, game.Player1, first)
				if !e.applyAction(g, game.Player1, action.CastSpell(earlier, nil, 0, nil)) {
					t.Fatal("could not cast first spell for storm")
				}
				e.resolveTopOfStack(g, log)
				def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
				cardID = addCardToHand(g, game.Player1, def)
				applied = e.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, log)
			}
			if !applied {
				t.Fatal("assistable cast was rejected")
			}
			obj, _ := g.Stack.Peek()
			wantCastCopy := route == "cast card copy" || route == "prepared copy"
			wantManaPayment := wantSpent > 0
			if obj.CastCopy != wantCastCopy || obj.ManaSpentToCast != wantSpent ||
				vault.Tapped != wantManaPayment || plateau.Tapped != wantManaPayment ||
				g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != 0 || spellCastEvents(g, obj.ID) != 1 {
				t.Fatalf("wrong paid cast or event: %+v", obj)
			}
			if route == "cast card copy" && lastSpellCastEvent(t, g).CascadeInstances != opt.Val(1) {
				t.Fatal("assisted card copy did not carry its true cast into cascade")
			}
			if route == "stack copies and storm" {
				copy := copyStackObject(e, g, obj)
				e.resolveTopOfStack(g, log)
				if copy == nil || !copy.Copy || copy.CastCopy || copy.ManaSpentToCast != wantSpent ||
					!e.putTriggeredAbilitiesOnStack(g) || !counterStackObject(g, obj.ID) {
					t.Fatal("expected ordinary copy and a storm trigger independent of the countered original")
				}
				e.resolveTopOfStack(g, log)
				stormCopy, _ := g.Stack.Peek()
				if stormCopy == nil || !stormCopy.Copy || stormCopy.CastCopy || stormCopy.ManaSpentToCast != wantSpent ||
					g.Stack.Size() != 1 || spellCastEventCount(g) != 2 {
					t.Fatal("stack copies wrongly counted as casts or storm lost its original")
				}
				e.resolveTopOfStack(g, log)
				if !p.Graveyard.Contains(cardID) || g.Players[game.Player2].ManaPool.Total() != 0 {
					t.Fatal("countering the original refunded Assist or moved a stack copy like a card")
				}
				return
			}
			e.resolveTopOfStack(g, log)
			switch route {
			case "cast card copy":
				if !p.Exile.Contains(cardID) {
					t.Fatal("cast card copy moved its original")
				}
			case "prepared copy":
				if permanentForCard(g, cardID) == nil {
					t.Fatal("prepared copy moved the original permanent")
				}
			default:
				if !p.Graveyard.Contains(cardID) {
					t.Fatal("card did not finish in its owner's graveyard")
				}
			}
		})
	}
}
