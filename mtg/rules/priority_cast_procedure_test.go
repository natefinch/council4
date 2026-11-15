package rules

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Official CR effective 2026-11-06: 601.2a-i/601.3, 115.5,
// 702.47a-e, 702.140a-f, 708.4-6, 702.37a-g, 401.5 and 734.1-2.
func priorityTestSource(g *game.Game, card id.ID, from zone.Type) {
	if from == zone.Hand {
		return
	}
	g.Players[game.Player1].Hand.Remove(card)
	destination, _ := destinationZone(g, game.Player1, from)
	destination.Add(card)
	if from == zone.Command {
		g.CommanderIDs[card] = true
	} else if from == zone.Library {
		castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
	} else {
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			Kind: game.RuleEffectCastFromZone, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
			CastFromZone: from, AffectedCardID: card, Duration: game.DurationPermanent,
		})
	}
}

func TestPriorityDismemberAnnouncedOnSameStackObject(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Graveyard, zone.Exile, zone.Library} {
		t.Run(from.String(), func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			bears := addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
			card := addCardToHand(g, game.Player1, registryCardDef(t, "Dismember"))
			priorityTestSource(g, card, from)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			var announced *game.StackObject
			var kinds []game.ChoiceKind
			agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
				top, ok := g.Stack.Peek()
				if !ok || g.Stack.Size() != 1 || top.SourceID != card || !top.Casting ||
					castSourceContains(g.Players[game.Player1], card, from) {
					t.Fatal("601.2a did not precede an announcement callback")
				}
				if announced != nil && top != announced {
					t.Fatal("casting replaced the announced stack object")
				}
				announced = top
				if stack := obs.Stack(); len(stack) != 1 || stack[0].ID != top.ID {
					t.Fatal("observer cannot see the announced spell")
				}
				kinds = append(kinds, request.Kind)
				for _, option := range request.Options {
					if option.Label == "Pay 2 life" || slices.Equal(option.Targets, []game.Target{game.PermanentTarget(bears.ObjectID)}) {
						return []int{option.Index}
					}
				}
				return request.DefaultSelection
			}}
			cast := action.CastSpellFaceFromZone(card, from, game.FaceFront, []game.Target{game.PermanentTarget(bears.ObjectID)}, 0, nil)
			if !actionsContain(e.legalActions(g, game.Player1), cast) ||
				!e.applyActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}) {
				t.Fatal("payable Dismember was not offered and cast")
			}
			top, _ := g.Stack.Peek()
			if !slices.Equal(kinds, []game.ChoiceKind{game.ChoicePayment, game.ChoicePayment, game.ChoiceTarget}) ||
				top != announced || top.Casting || top.ManaSpentToCast != 1 ||
				g.Players[game.Player1].Life != 36 || g.Players[game.Player1].ManaPool.Total() != 0 || spellCastEventCount(g) != 1 {
				t.Fatalf("kinds=%v life=%d pool=%d: want two life announcements, target, and one same-object cast",
					kinds, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total())
			}
			e.resolveTopOfStack(g, &TurnLog{})
			e.applyStateBasedActions(g)
			if _, ok := permanentByObjectID(g, bears.ObjectID); ok {
				t.Fatal("Dismember did not resolve its -5/-5 on the Bears")
			}
		})
	}
}

func TestPriorityMorphLandUsesFaceDownPermissionCharacteristics(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	card := addCardToLibrary(g, game.Player1, registryCardDef(t, "Zoetic Cavern"))
	castSpellsFromLibraryTopPermanent(g, game.Player1, []types.Card{types.Creature})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	cast := action.CastFaceDownFromZone(card, zone.Library, game.FaceFront, game.FaceDownMorph)
	if !actionsContain(e.legalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) {
		t.Fatal("the printed land was rejected instead of checking its face-down creature characteristics (CR 708.4)")
	}
	requireFaceDownSpell(t, g, card, game.FaceDownMorph, zone.Library, 3)
}

func TestPriorityMentalMisstepCannotTargetItself(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	card := addCardToHand(g, game.Player1, registryCardDef(t, "Mental Misstep"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	if slices.ContainsFunc(e.legalActions(g, game.Player1), func(act action.Action) bool {
		cast, ok := act.CastSpellPayload()
		return ok && cast.CardID == card
	}) {
		t.Fatal("Mental Misstep offered with itself as its only possible target (CR 115.5)")
	}
	other := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
		SourceID: addCardToHand(g, game.Player2, registryCardDef(t, "Sol Ring")), Controller: game.Player2}
	g.Stack.Push(other)
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceTarget {
			if len(request.Options) != 1 || !slices.Equal(request.Options[0].Targets, []game.Target{game.StackObjectTarget(other.ID)}) {
				t.Fatal("casting Mental Misstep was offered as its own target")
			}
		}
		return request.DefaultSelection
	}}
	if !e.applyActionWithChoices(g, game.Player1, action.CastSpell(card, []game.Target{game.StackObjectTarget(other.ID)}, 0, nil),
		reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("Mental Misstep could not target another one-mana spell")
	}
}

func TestPriorityPhyrexianOptionBeyondTargetSample(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		t.Run(map[bool]string{false: "hand", true: "prepared copy"}[prepared], func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			def := *registryCardDef(t, "Dismember")
			def.ManaCost = opt.Val(append(cost.Mana{cost.X}, def.ManaCost.Val...))
			mode := def.SpellAbility.Val.Modes[0]
			mode.Targets = slices.Clone(mode.Targets)
			mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets = 2, 2
			def.SpellAbility = opt.Val(mode.Ability())
			for range 11 {
				addCombatPermanent(g, game.Player2, registryCardDef(t, "Elderwood Scion"))
			}
			var targets []game.Target
			for range 2 {
				targets = append(targets, game.PermanentTarget(addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears")).ObjectID))
			}
			card := id.ID(0)
			if prepared {
				token := preparedLargeTargetToken(t, g, &def)
				// Keep the prepared source out of this creature-target domain.
				token.TokenDef.Types = []types.Card{types.Artifact}
				card = token.ObjectID
			} else {
				card = addCardToHand(g, game.Player1, &def)
			}
			g.Players[game.Player1].ManaPool.Add(mana.B, 2)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			phyrexian, targetChoices := 0, 0
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoicePayment {
					phyrexian++
					for _, option := range request.Options {
						if option.Label == "Pay with mana" {
							return []int{option.Index}
						}
					}
					t.Fatal("mana announcement withheld despite the payable 78th target pair")
				}
				if request.Kind == game.ChoiceTarget {
					if phyrexian != 2 {
						t.Fatal("targets preceded the Phyrexian announcements")
					}
					for _, option := range request.Options {
						if !slices.Contains(targets, option.Targets[0]) {
							t.Fatal("an unpayable Scion was offered after binding mana intentions")
						}
					}
					targetChoices++
				}
				return request.DefaultSelection
			}}
			var offered action.Action
			for _, act := range e.legalActions(g, game.Player1) {
				if cast, ok := act.CastSpellPayload(); ok && cast.CardID == card && cast.XValue == 0 {
					offered = act
					break
				}
			}
			if offered.Kind != action.ActionCastSpell || !e.applyActionWithChoices(g, game.Player1, offered, reboundKickerAgents(agent), &TurnLog{}) {
				t.Fatal("the payable target pair beyond the former 64-sample was not cast")
			}
			top, _ := g.Stack.Peek()
			if phyrexian != 2 || targetChoices != 2 || !targetsContainAll(top.Targets, targets) ||
				top.ManaSpentToCast != 3 || g.Players[game.Player1].Life != 40 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("the exact two-Bears payment was not committed")
			}
		})
	}
}

func TestPriorityTargetCompletionWorkIsBounded(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	def := *registryCardDef(t, "Dismember")
	def.ManaCost = opt.Val(append(cost.Mana{cost.X}, def.ManaCost.Val...))
	mode := def.SpellAbility.Val.Modes[0]
	mode.Targets = slices.Clone(mode.Targets)
	mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets = 2, 2
	def.SpellAbility = opt.Val(mode.Ability())
	for range 11 {
		addCombatPermanent(g, game.Player2, registryCardDef(t, "Elderwood Scion"))
	}
	for range 2 {
		addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
	}
	card := addCardToHand(g, game.Player1, &def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	view := priorityCastView(g, &game.StackObject{
		Kind: game.StackSpell, SourceID: card, Controller: game.Player1, SourceZone: zone.Hand,
	})
	request := payment.SpellRequest{PlayerID: game.Player1, CardID: card, SourceZone: zone.Hand, Card: &def}
	complete := e.pendingSpellPayment(view, &def, action.CastSpellAction{})
	partial := payment.Partial{Phyrexian: map[int]bool{0: false, 1: false}}
	allocations := testing.AllocsPerRun(3, func() {
		plan, ok := complete(request, partial)
		if !ok || plan.Mana().Life != 0 || totalManaSpent(plan.Mana().PoolSpend) != 3 {
			t.Fatal("the payable 78th pair did not retain its exact mana-only witness")
		}
	})
	t.Logf("allocations for all 78 exact target pairs: %.0f", allocations)
	// Rebuilding the whole board's static rules for every symbol branch makes
	// these 78 payments allocate hundreds of thousands of intermediate effects.
	if allocations > 100_000 {
		t.Fatalf("exact target completion rebuilt static rules repeatedly: %.0f allocations", allocations)
	}
	if view.InStaticSourceFrame() || g.InStaticSourceFrame() {
		t.Fatal("target completion retained a read frame beyond its pure query")
	}
}

func TestPriorityXUsesPlannedManaInEveryZone(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Command, zone.Graveyard, zone.Exile, zone.Library, zone.Battlefield} {
		t.Run(from.String(), func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			card, face := id.ID(0), game.FaceFront
			if from == zone.Battlefield {
				card, face = preparedSpellToken(t, g, "Exsanguinate").ObjectID, game.FaceAlternate
			} else {
				card = addCardToHand(g, game.Player1, registryCardDef(t, "Exsanguinate"))
				priorityTestSource(g, card, from)
			}
			grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			g.Players[game.Player1].ManaPool.Add(mana.B, 2)
			cast := action.CastSpellFaceFromZone(card, from, face, nil, 1, nil)
			if !actionsContain(e.legalActions(g, game.Player1), cast) {
				t.Fatal("X=1 payable with the planned Grove was not offered")
			}
			if !e.applyAction(g, game.Player1, cast) {
				t.Fatal("the X=1 cast did not complete its witness")
			}
			top, _ := g.Stack.Peek()
			if top.XValue != 1 || top.ManaSpentToCast != 3 || !grove.Tapped ||
				g.Players[game.Player1].Life != 39 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("X did not bind the exact planned mana payment")
			}
		})
	}
}

func TestPriorityManaEffectsEnableNonmanaCosts(t *testing.T) {
	for _, kind := range []string{"counter", "life", "conflicting counter"} {
		t.Run(kind, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			spell := *registryCardDef(t, "Sol Ring")
			body := game.TapManaAbility(mana.C)
			body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence,
				game.Instruction{Primitive: game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(1)}})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Counter mana source", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{body},
			}})
			spell.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalRemoveCounterAmong, CounterKind: counter.Charge, Amount: 1,
				MatchPermanentType: true, PermanentType: types.Artifact}}
			if kind == "life" {
				g.Battlefield = nil
				source = addCombatPermanent(g, game.Player1, registryCardDef(t, "The Great Henge"))
				spell.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 42}}
			} else if kind == "conflicting counter" {
				source.Counters.Add(counter.Charge, 1)
				body.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}
				body.Content = game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.C, Amount: game.Fixed(1)}}}}.Ability()
				g.CardInstances[source.CardInstanceID].Def.ManaAbilities = []game.ManaAbility{body}
			}
			card := addCardToHand(g, game.Player1, &spell)
			cast := action.CastSpell(card, nil, 0, nil)
			offered := actionsContain(e.legalActions(g, game.Player1), cast)
			if kind == "conflicting counter" {
				if offered || e.applyAction(g, game.Player1, cast) || source.Counters.Get(counter.Charge) != 1 {
					t.Fatal("the same charge counter was reserved for mana and the spell")
				}
				return
			}
			if !offered || !e.applyAction(g, game.Player1, cast) {
				t.Fatal("a valid payment-time mana effect did not enable the spell's nonmana cost")
			}
			if !source.Tapped || !source.Counters.IsEmpty() || kind == "life" && g.Players[game.Player1].Life != 0 {
				t.Fatal("the enabled nonmana cost was not paid")
			}
		})
	}
}

func TestPriorityMutateAnnouncedOnSameStackObject(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Command, zone.Graveyard, zone.Exile, zone.Library} {
		for _, variable := range []bool{false, true} {
			t.Run(from.String()+map[bool]string{false: "/registered", true: "/X and Phyrexian"}[variable], func(t *testing.T) {
				e, g := newEffectiveCascadeGame()
				def := *registryCardDef(t, "Parcelbeast")
				if variable {
					def.StaticAbilities = slices.Clone(def.StaticAbilities)
					def.StaticAbilities[0].KeywordAbilities = []game.KeywordAbility{
						game.MutateKeyword{Cost: cost.Mana{cost.X, cost.PhyrexianMana(mana.G)}},
					}
				}
				target := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
				grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
				card := addCardToHand(g, game.Player1, &def)
				priorityTestSource(g, card, from)
				cast, _ := actionBuild.castMutateSpell(card, from, target.ObjectID).CastSpellPayload()
				if variable {
					cast.XValue = 1
				} else {
					g.Players[game.Player1].ManaPool.Add(mana.G, 1)
				}
				var announced *game.StackObject
				phyrexian, targets := 0, 0
				agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
					top, ok := g.Stack.Peek()
					if !ok || !top.Casting || top.SourceID != card || len(obs.Stack()) != 1 ||
						castSourceContains(g.Players[game.Player1], card, from) || announced != nil && top != announced {
						t.Fatal("mutate did not preserve the 601.2a object during announcements")
					}
					announced = top
					if request.Kind == game.ChoiceTarget {
						targets++
						if variable && phyrexian != 1 {
							t.Fatal("mutate target preceded the Phyrexian intention")
						}
					}
					for _, option := range request.Options {
						if option.Label == "Pay 2 life" {
							phyrexian++
							return []int{option.Index}
						}
						if option.Label == "U" || slices.Equal(option.Targets, []game.Target{game.PermanentTarget(target.ObjectID)}) {
							return []int{option.Index}
						}
					}
					return request.DefaultSelection
				}}
				act := action.AnnounceSpell(cast)
				if !actionsContain(e.legalActions(g, game.Player1), act) ||
					!e.applyActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}) {
					t.Fatal("the witnessed mutate route was not offered and cast")
				}
				top, _ := g.Stack.Peek()
				spent, life := 2, 39
				if variable {
					spent, life = 1, 37
				}
				if top != announced || top.Casting || !top.Mutate || top.XValue != cast.XValue ||
					top.MutateTargetID != target.ObjectID || targets != 1 || !grove.Tapped ||
					top.ManaSpentToCast != spent || g.Players[game.Player1].Life != life {
					t.Fatal("mutate lost an announcement, target or exact payment")
				}
			})
		}
	}
}

func TestPriorityFaceDownPermissionZones(t *testing.T) {
	for _, name := range []string{"War Behemoth", "Exit Specialist"} {
		for _, from := range []zone.Type{zone.Hand, zone.Command, zone.Graveyard, zone.Exile, zone.Library} {
			t.Run(name+"/"+from.String(), func(t *testing.T) {
				e, g := newEffectiveCascadeGame()
				card := addCardToHand(g, game.Player1, registryCardDef(t, name))
				priorityTestSource(g, card, from)
				grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
				g.Players[game.Player1].ManaPool.Add(mana.C, 2)
				kind := game.FaceDownMorph
				if name == "Exit Specialist" {
					kind = game.FaceDownDisguise
				}
				act := action.CastFaceDownFromZone(card, from, game.FaceFront, kind)
				if from == zone.Hand {
					act = actionBuild.castFaceDown(card, game.FaceFront, kind)
				}
				var announced *game.StackObject
				agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
					top, ok := g.Stack.Peek()
					if !ok || !top.FaceDown || !top.Casting || top.SourceID != card ||
						announced != nil && announced != top || len(obs.Stack()) != 1 ||
						castSourceContains(g.Players[game.Player1], card, from) {
						t.Fatal("face-down payment did not use its announced stack object")
					}
					announced = top
					return request.DefaultSelection
				}}
				if !actionsContain(e.legalActions(g, game.Player1), act) ||
					!e.applyActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}) {
					t.Fatal("the authorized face-down route was not offered and cast")
				}
				top, _ := g.Stack.Peek()
				if top != announced || top.Casting || !grove.Tapped || g.Players[game.Player1].Life != 39 {
					t.Fatal("face-down casting did not commit its witnessed payment")
				}
				requireFaceDownSpell(t, g, card, kind, from, 3)
			})
		}
	}
}

func TestPrioritySpliceRevealsAllBeforeAnyTargets(t *testing.T) {
	for _, hostName := range []string{"Desperate Ritual", "Glacial Ray"} {
		t.Run(hostName, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			host := addCardToHand(g, game.Player1, registryCardDef(t, hostName))
			ritual := addCardToHand(g, game.Player1, registryCardDef(t, "Desperate Ritual"))
			ray := addCardToHand(g, game.Player1, registryCardDef(t, "Glacial Ray"))
			target := addCombatPermanent(g, game.Player2, registryCardDef(t, "Blightsteel Colossus"))
			g.Players[game.Player1].ManaPool.Add(mana.R, 3)
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			remaining := []id.ID{ritual, ray}
			targets := 0
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoiceZoneSelection {
					if len(remaining) == 0 {
						return nil
					}
					for _, option := range request.Options {
						if option.Card.Exists && option.Card.Val.CardID == remaining[0] {
							remaining = remaining[1:]
							return []int{option.Index}
						}
					}
				}
				if request.Kind == game.ChoiceTarget {
					top, _ := g.Stack.Peek()
					shown := NewObservation(g, game.Player2).AnnouncedCards()
					if len(remaining) != 0 || top.SourceID != host || !top.Casting || spellCastEventCount(g) != 0 ||
						len(shown) != 2 || shown[0].CardInstanceID != ritual || shown[1].CardInstanceID != ray {
						t.Fatal("an observer did not see both splice cards before the first target")
					}
					targets++
					for _, option := range request.Options {
						if slices.Equal(option.Targets, []game.Target{game.PermanentTarget(target.ObjectID)}) {
							return []int{option.Index}
						}
					}
				}
				return request.DefaultSelection
			}}
			var hostTargets []game.Target
			if hostName == "Glacial Ray" {
				hostTargets = []game.Target{game.PermanentTarget(target.ObjectID)}
			}
			if !e.applyActionWithChoices(g, game.Player1, action.CastSpell(host, hostTargets, 0, nil),
				reboundKickerAgents(agent), &TurnLog{}) {
				t.Fatal("registered host and both splice cards did not complete")
			}
			top, _ := g.Stack.Peek()
			expected := 1 + len(hostTargets)
			if targets != expected || len(top.SplicedContent) != 2 || !g.Players[game.Player1].Hand.Contains(ritual) ||
				!g.Players[game.Player1].Hand.Contains(ray) || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("the combined splice announcement or payment was lost")
			}
			e.resolveTopOfStack(g, &TurnLog{})
			if target.MarkedDamage != 2*expected || g.Players[game.Player1].ManaPool.Amount(mana.R) != 3*(2-len(hostTargets)) {
				t.Fatal("the host and all spliced effects did not resolve")
			}
		})
	}
}

func TestPriorityPhyrexianKickerBindsBeforeTargets(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	def := *registryCardDef(t, "Gitaxian Probe")
	def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{
		game.KickerKeyword{Cost: cost.Mana{cost.PhyrexianMana(mana.B)}},
	}}}
	card := addCardToHand(g, game.Player1, &def)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	phyrexian := 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoicePayment && strings.HasPrefix(request.Prompt, "Pay {") {
			phyrexian++
			if phyrexian == 2 && (len(request.Options) != 1 || request.Options[0].Label != "Pay 2 life") {
				t.Fatal("the kicker offered a mana intention without a completion")
			}
		}
		if request.Kind == game.ChoiceTarget && phyrexian != 2 {
			t.Fatal("the kicked Probe chose targets before both Phyrexian intentions")
		}
		return request.DefaultSelection
	}}
	act := action.CastKickedSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)
	if !e.applyActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatal("the kicked Probe did not complete")
	}
	top, _ := g.Stack.Peek()
	if !top.KickerPaid || phyrexian != 2 || g.Players[game.Player1].Life != 38 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("the mana announcement fell back to life or the kicker intention was not honored")
	}
}

func TestPriorityEscapeRejectsLibraryMovingPaymentAbility(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Underworld Breach"))
	card := addCardToGraveyard(g, game.Player1, registryCardDef(t, "Sol Ring"))
	for range 2 {
		addCardToGraveyard(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	}
	millikin := addCombatPermanent(g, game.Player1, registryCardDef(t, "Millikin"))
	milled := addCardToLibrary(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	cast := action.CastSpellFaceFromZone(card, zone.Graveyard, game.FaceFront, nil, 0, nil)
	if actionsContain(e.legalActions(g, game.Player1), cast) || e.applyAction(g, game.Player1, cast) ||
		millikin.Tapped || !g.Players[game.Player1].Library.Contains(milled) {
		t.Fatal("escape was offered as though Millikin could activate during payment (CR 605.1a)")
	}
	if !e.applyAction(g, game.Player1, sourceActivation(t, e, g, millikin.ObjectID)) {
		t.Fatal("Millikin could not be activated with priority")
	}
	if top, _ := g.Stack.Peek(); top.Kind != game.StackActivatedAbility || g.Players[game.Player1].ManaPool.Total() != 0 ||
		!g.Players[game.Player1].Graveyard.Contains(milled) {
		t.Fatal("Millikin did not use the stack with its mill paid as an activation cost")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !actionsContain(e.legalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) ||
		g.Players[game.Player1].Exile.Size() != 3 || g.Players[game.Player1].Graveyard.Size() != 0 ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("standalone Millikin resolution did not enable the now-legal escape")
	}
}

func TestPriorityExternalTargetsRetainCompletion(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	def := *registryCardDef(t, "Dismember")
	mode := def.SpellAbility.Val.Modes[0]
	mode.Targets = slices.Clone(mode.Targets)
	mode.Targets[0].Chooser = game.TargetChooserOpponent
	def.SpellAbility = opt.Val(mode.Ability())
	addCombatPermanent(g, game.Player2, registryCardDef(t, "Elderwood Scion"))
	bears := addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
	card := addCardToHand(g, game.Player1, &def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	announcements, targets := 0, 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoicePayment {
			announcements++
		}
		if request.Kind == game.ChoiceTarget {
			targets++
			if announcements != 2 || len(request.Options) != 1 ||
				!slices.Equal(request.Options[0].Targets, []game.Target{game.PermanentTarget(bears.ObjectID)}) {
				t.Fatal("the opponent was offered a target without a joint payment witness")
			}
		}
		return request.DefaultSelection
	}}
	agents := [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}
	if !e.applyActionWithChoices(g, game.Player1, action.CastSpell(card, []game.Target{game.DeferredTarget()}, 0, nil),
		agents, &TurnLog{}) {
		t.Fatal("the externally chosen target did not complete")
	}
	top, _ := g.Stack.Peek()
	if targets != 1 || top.Targets[0].PermanentID != bears.ObjectID || g.Players[game.Player1].Life != 36 {
		t.Fatal("the external target or binding payment was lost")
	}
}

func TestPriorityStrixChoosesAdditionalAltarOccurrence(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	altar := addCombatPermanent(g, game.Player1, registryCardDef(t, "Phyrexian Altar"))
	first := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	second := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	island := addCombatPermanent(g, game.Player1, registryCardDef(t, "Island"))
	strix := addCardToHand(g, game.Player1, registryCardDef(t, "Baleful Strix"))
	draw := addCardToLibrary(g, game.Player1, registryCardDef(t, "Forest"))
	colors, selections := []string{"B", "U"}, 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Activate a mana ability while paying?" {
			if selections == len(colors) {
				return request.DefaultSelection
			}
			for _, option := range request.Options {
				if option.SourcePermanent == altar.ObjectID {
					selections++
					return []int{option.Index}
				}
			}
			t.Fatal("the next Altar occurrence was withheld despite a complete two-sacrifice payment")
		}
		if request.Prompt == "Sacrifice a creature" {
			if selections != 1 {
				t.Fatal("the second occurrence offered the already-sacrificed first Bear")
			}
			for _, option := range request.Options {
				if slices.Equal(option.Targets, []game.Target{game.PermanentTarget(first.ObjectID)}) {
					return []int{option.Index}
				}
			}
			t.Fatal("the first Bear was not offered as this occurrence's cost")
		}
		if request.Prompt == "Choose a color" {
			for _, option := range request.Options {
				if option.Label == colors[selections-1] {
					return []int{option.Index}
				}
			}
			t.Fatal("the occurrence's chosen color was not offered")
		}
		return request.DefaultSelection
	}}
	cast, log := action.CastSpell(strix, nil, 0, nil), &TurnLog{}
	if !actionsContain(e.legalActions(g, game.Player1), cast) ||
		!e.applyActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), log) {
		t.Fatal("Baleful Strix did not complete its chosen payment")
	}
	obj, _ := g.Stack.Peek()
	if selections != 2 || island.Tapped || obj.SourceID != strix || obj.ManaSpentToCast != 2 ||
		g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player1].Life != 40 ||
		!g.Players[game.Player1].Graveyard.Contains(first.CardInstanceID) ||
		!g.Players[game.Player1].Graveyard.Contains(second.CardInstanceID) {
		t.Fatal("the chosen two-Altar payment did not preserve Island and pay both distinct sacrifices")
	}
	var produced []mana.Color
	for _, event := range g.Events {
		if event.Kind == game.EventManaProduced && event.SourceObjectID == altar.ObjectID {
			produced = append(produced, event.ProducedManaColors...)
		}
	}
	if !slices.Equal(produced, []mana.Color{mana.B, mana.U}) {
		t.Fatalf("Altar produced %v, want independent black then blue", produced)
	}
	requirePaymentManaActivationsLogged(t, log, altar.ObjectID, altar.ObjectID)
	e.resolveTopOfStack(g, log)
	e.putTriggeredAbilitiesOnStack(g)
	e.resolveTopOfStack(g, log)
	if !g.Players[game.Player1].Hand.Contains(draw) || g.Stack.Size() != 0 {
		t.Fatal("Baleful Strix did not resolve and draw its card")
	}
}

func TestPriorityOpaqueFailureRestoresCastSourceAndCopy(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Graveyard, zone.Exile, zone.Library, zone.Battlefield} {
		t.Run(from.String(), func(t *testing.T) {
			g, e, _, _ := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
			card, face := id.ID(0), game.FaceFront
			if from == zone.Battlefield {
				card, face = phyrexianPreparedToken(t, g).ObjectID, game.FaceAlternate
			} else {
				card = addCardToHand(g, game.Player1, registryCardDef(t, "Gitaxian Probe"))
				priorityTestSource(g, card, from)
			}
			before, log := g.Clone(), &TurnLog{}
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				for _, option := range request.Options {
					if option.Label == "Pay with mana" || strings.Contains(option.Label, "(opaque)") {
						return []int{option.Index}
					}
				}
				return request.DefaultSelection
			}}
			outcome, _ := e.attemptActionWithChoices(g, game.Player1,
				action.CastSpellFaceFromZone(card, from, face, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil),
				reboundKickerAgents(agent), log)
			if outcome != actionReversed || !g.Stack.IsEmpty() || g.IDGen.Current() != before.IDGen.Current() ||
				!reflect.DeepEqual(g.Players, before.Players) || !reflect.DeepEqual(g.CardInstances, before.CardInstances) ||
				!reflect.DeepEqual(g.Battlefield, before.Battlefield) || !reflect.DeepEqual(g.Events, before.Events) ||
				len(log.Entries) != 0 || spellCastEventCount(g) != 0 {
				t.Fatal("opaque failure did not restore the world and cancel its cast-source transition (CR 734.1)")
			}
		})
	}
}
