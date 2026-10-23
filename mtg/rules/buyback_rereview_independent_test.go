package rules

import (
	"fmt"
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
	"github.com/natefinch/council4/opt"
)

// Fresh repair-round probes at f2c7f5c90c4d38aef2271b5ed81f639e4a55188d.
// Primary sources read: verified CR 2026-09-25, 118.8/9, 400.7,
// 601.2a-i, 702.27a, 702.85a, 707.10/12, 614.15, 616.1; exact Reiterate printing
// 99302c41-434f-41ff-a61a-2c3681a0c135 and all seven WotC 2021-03-19
// rulings. Variations of the unregistered fixture are Engine-only probes.
func TestBuybackRereviewBoundCasterAndDistinctPayments(t *testing.T) {
	for _, cards := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("caster-hand-%d", cards), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			def := reiterateFixture()
			def.AdditionalCosts = []cost.Additional{{
				Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "mandatory",
			}}
			def.StaticAbilities[0].KeywordAbilities = []game.KeywordAbility{game.BuybackKeyword{
				Cost: cost.Mana{cost.O(3)},
				AdditionalCosts: []cost.Additional{{
					Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "buyback",
				}},
			}}
			cardID := addCardToHand(g, game.Player4, def)
			if !moveCardBetweenZones(g, game.Player4, cardID, zone.Hand, zone.Exile) {
				t.Fatal("could not exile source card")
			}
			version := g.CardInstances[cardID].ZoneVersion
			g.AddExileCounterFromController(cardID, counter.Void, 2, game.Player4)
			source := addCombatPermanent(g, game.Player1, isochronScepterDef())
			obj := activatedObjFor(source)
			imprintLink(g, obj, cardID)
			otherHand := addCardToHand(g, game.Player1, vanillaCreature("Wrong payer", 1, 1))
			var discards []id.ID
			for range cards {
				discards = append(discards, addCardToHand(g, game.Player3, vanillaCreature("Caster payment", 1, 1)))
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, 9)
			g.Players[game.Player3].ManaPool.Add(mana.C, 4)
			g.CostModifiers = []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 1}}
			offers := 0
			agents := [game.NumPlayers]PlayerAgent{game.Player3: &authoredCopyChoiceAgent{
				choose: func(request game.ChoiceRequest) []int {
					if request.Player != game.Player3 {
						t.Fatal("cost decision sent to someone other than bound caster")
					}
					if request.Kind == game.ChoicePayment {
						for _, option := range request.Options {
							if strings.Contains(option.Label, "Buyback") {
								offers++
								return []int{option.Index}
							}
						}
					}
					return request.DefaultSelection
				},
			}}
			content := game.Mode{Sequence: []game.Instruction{{
				Primitive: game.PlayLinkedExiledCard{
					Player: game.CapturedPlayerReference(game.Player3), LinkID: testImprintLink,
					Copy: true, WithoutPayingManaCost: true,
				},
			}}}.Ability()
			beforeEvents := len(g.Events)
			engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if !g.Players[game.Player4].Exile.Contains(cardID) ||
				g.CardInstances[cardID].ZoneVersion != version ||
				g.ExileCounterCount(cardID, counter.Void) != 2 ||
				!g.Players[game.Player1].Hand.Contains(otherHand) ||
				g.Players[game.Player1].ManaPool.Total() != 9 {
				t.Fatal("copy cast mutated the original card or used the source controller's resources")
			}
			if cards == 0 {
				if g.Stack.Size() != 1 || g.Players[game.Player3].ManaPool.Total() != 4 ||
					len(g.Events) != beforeEvents || offers != 0 {
					t.Fatal("unpayable mandatory discard partially cast or paid")
				}
				return
			}
			copied, _ := g.Stack.Peek()
			paid := cards == 2
			wantSpent, wantOffers := 1, 0
			if paid {
				wantSpent, wantOffers = 4, 1
			}
			if copied.Controller != game.Player3 || !copied.Copy ||
				!copied.CopyOwner.Exists || copied.CopyOwner.Val != game.Player3 ||
				copied.BuybackPaid != paid || copied.ManaSpentToCast != wantSpent ||
				g.Players[game.Player3].ManaPool.Total() != 4-wantSpent ||
				offers != wantOffers || len(copied.PaidCostSubjects) != cards ||
				len(copied.Targets) != 1 || copied.Targets[0] != game.StackObjectTarget(subject.ID) {
				t.Fatalf("wrong bound cast/payment/target facts: copy=%+v offers=%d", copied, offers)
			}
			seen := make(map[id.ID]bool)
			for _, payment := range copied.PaidCostSubjects {
				if seen[payment.Snapshot.CardID] {
					t.Fatal("one card paid both mandatory and buyback discard")
				}
				seen[payment.Snapshot.CardID] = true
			}
			for _, discarded := range discards {
				if !seen[discarded] || !g.Players[game.Player3].Graveyard.Contains(discarded) {
					t.Fatal("payment did not retain distinct actual caster-owned discard subjects")
				}
			}
			for _, branch := range []*game.Game{g, g.Clone()} {
				engine.resolveTopOfStack(branch, &TurnLog{})
				engine.resolveTopOfStack(branch, &TurnLog{})
				if branch.Players[game.Player3].Life != 44 || branch.Players[game.Player1].Life != 40 ||
					!branch.Players[game.Player4].Exile.Contains(cardID) ||
					branch.CardInstances[cardID].ZoneVersion != version ||
					branch.ExileCounterCount(cardID, counter.Void) != 2 {
					t.Fatal("copy-relative controller or original exile identity changed on resolution")
				}
			}
		})
	}
}

func TestBuybackRereviewCastCopyStormCascadeAndRiders(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	castBuybackSubject(t, g, engine)
	def := reiterateFixture()
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		KeywordAbilities: game.SimpleKeywords(game.Storm, game.Cascade),
	})
	cardID := addCardToHand(g, game.Player4, def)
	if !moveCardBetweenZones(g, game.Player4, cardID, zone.Hand, zone.Exile) {
		t.Fatal("could not exile source card")
	}
	version := g.CardInstances[cardID].ZoneVersion
	g.AddExileCounterFromController(cardID, counter.Void, 1, game.Player4)
	source := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(source)
	imprintLink(g, obj, cardID)
	hit := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Independent cascade life", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
		}}}.Ability()),
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	addRiders(g, game.Player1, mana.C, 3)
	t.Run("physical-free-cast-control", func(t *testing.T) {
		control := g.Clone()
		if !engine.castFreeOption(control, control.Players[game.Player4], game.Player1,
			control.CardInstances[cardID], action.CastSpellAction{
				CardID: cardID, SourceZone: zone.Exile,
				Targets: []game.Target{game.StackObjectTarget(control.Stack.Objects()[0].ID)},
			}, false, independentBuybackAgent(t, independentChooseBuyback), &TurnLog{}) {
			t.Fatal("otherwise equivalent physical free cast failed")
		}
		controlObjects := control.Stack.Objects()
		if len(controlObjects) != 4 || controlObjects[3].SourceID != hit ||
			controlObjects[1].ManaSpentToCast != 3 || !controlObjects[1].BuybackPaid ||
			!controlObjects[2].BuybackPaid || control.Players[game.Player1].ManaPool.Total() != 0 ||
			len(control.Players[game.Player1].ManaRiders) != 0 {
			t.Fatal("physical free-cast control did not cast its cascade hit, pay buyback, and consume riders")
		}
	})
	beforeEvents := len(g.Events)
	log := &TurnLog{}
	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(),
		independentBuybackAgent(t, independentChooseBuyback), log)
	objects := g.Stack.Objects()
	if len(objects) != 4 || !objects[1].Copy || !objects[1].BuybackPaid ||
		!objects[2].Copy || !objects[2].BuybackPaid || objects[3].SourceID != hit ||
		objects[1].ManaSpentToCast != 3 || g.Players[game.Player1].ManaPool.Total() != 0 ||
		len(g.Players[game.Player1].ManaRiders) != 0 || firedSpendRiderCount(g) != 0 {
		for _, object := range objects {
			t.Logf("stack: id=%v source=%v copy=%v paid=%v spent=%v", object.ID, object.SourceID, object.Copy, object.BuybackPaid, object.ManaSpentToCast)
		}
		for _, choice := range log.Choices {
			t.Logf("choice: player=%v prompt=%q selected=%v", choice.Request.Player, choice.Request.Prompt, choice.Selected)
		}
		t.Fatalf("cast copy skipped or duplicated storm/cascade/rider handling: casts=%d copies=%d library=%v riders=%d fired=%d",
			countEventsOfKind(g.Events[beforeEvents:], game.EventSpellCast),
			countEventsOfKind(g.Events[beforeEvents:], game.EventSpellCopied),
			g.Players[game.Player1].Library.All(), len(g.Players[game.Player1].ManaRiders), firedSpendRiderCount(g))
	}
	if countEventsOfKind(g.Events[beforeEvents:], game.EventSpellCast) != 2 ||
		countEventsOfKind(g.Events[beforeEvents:], game.EventSpellCopied) != 1 {
		t.Fatal("cast card copy or its non-cast storm copy produced wrong cast/copy history")
	}
	for _, event := range g.Events[beforeEvents:] {
		if event.CardID == cardID &&
			(event.Kind == game.EventZoneChanged || event.Kind == game.EventCardPlayedFromExile) {
			t.Fatal("copy reported the original physical card leaving exile")
		}
		if event.Kind == game.EventSpellCast && event.StackObjectID == objects[1].ID &&
			(event.CardID != 0 || event.TokenDef == nil || event.Controller != game.Player1 ||
				!event.ManaSpentToCast.Exists || event.ManaSpentToCast.Val != 3 ||
				event.PlayerEventOrdinalThisTurn != 1) {
			t.Fatalf("cast card copy has incorrect event identity or paid facts: %+v", event)
		}
	}
	for _, branch := range []*game.Game{g, g.Clone()} {
		for range 8 {
			if branch.Stack.Size() == 0 {
				break
			}
			engine.resolveTopOfStack(branch, &TurnLog{})
		}
		if branch.Stack.Size() != 0 || branch.Players[game.Player1].Life != 50 ||
			branch.Players[game.Player2].Life != 44 ||
			!branch.Players[game.Player1].Graveyard.Contains(hit) ||
			!branch.Players[game.Player4].Exile.Contains(cardID) ||
			branch.CardInstances[cardID].ZoneVersion != version ||
			branch.ExileCounterCount(cardID, counter.Void) != 1 {
			t.Fatal("cast copy, storm copy or clone resolved against the wrong source/controller")
		}
	}
}

func TestBuybackRereviewSacrificedImprintSourceDoesNotInvalidateCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	castBuybackSubject(t, g, engine)
	def := reiterateFixture()
	def.AdditionalCosts = []cost.Additional{{
		Kind: cost.AdditionalSacrifice, Amount: 1, MatchPermanentType: true,
		PermanentType: types.Artifact, SubjectKey: "source-sacrifice",
	}}
	cardID := addExiledInstant(g, game.Player4, def)
	g.AddExileCounterFromController(cardID, counter.Void, 1, game.Player4)
	source := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(source)
	imprintLink(g, obj, cardID)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(),
		independentBuybackAgent(t, independentChooseBuyback), &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || g.Stack.Size() != 2 || !copied.BuybackPaid || !copied.Copy ||
		!g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) ||
		len(copied.PaidCostSubjects) != 1 ||
		copied.PaidCostSubjects[0].Snapshot.CardID != source.CardInstanceID {
		t.Fatal("copy could not pay its imprint source as an ordinary additional sacrifice")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 44 || !g.Players[game.Player4].Exile.Contains(cardID) ||
		g.CardInstances[cardID].ZoneVersion != 0 || g.ExileCounterCount(cardID, counter.Void) != 1 {
		t.Fatal("paid cast copy lost its body or changed the original after its imprint source left")
	}
}
