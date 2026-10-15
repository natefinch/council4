package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestAuthoredSearchSameLibraryReplacementHasNoArrivalReceipt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToLibrary(g, game.Player1, selfLibraryReplacementCardDef())
	card, _ := g.GetCardInstance(cardID)
	version := card.ZoneVersion
	epoch := g.Players[game.Player1].Library.ReorderVersion()
	ctx := &programContext{
		engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player1},
		active: true, log: &TurnLog{},
	}
	found, err := ctx.Search(game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(1),
		Spec: game.SearchSpec{SourceZone: zone.Library, RevealOnly: true, Reveal: true},
	})
	if err != nil || len(found.Cards) != 1 {
		t.Fatalf("revealed search: %+v, %v", found, err)
	}
	arrivals, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Graveyard})
	if err != nil || len(arrivals) != 0 || !g.Players[game.Player1].Library.Contains(cardID) ||
		card.ZoneVersion != version || g.Players[game.Player1].Library.ReorderVersion() != epoch+1 {
		t.Fatalf("same-library replacement invented an arrival or missed its shuffle: %+v, %v", arrivals, err)
	}
}

func TestAuthoredDestroyCorrelatesFrozenBeforeAndActualArrivals(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 2)
	obj, _ := g.Stack.Peek()
	card, _ := g.GetCardInstance(targets[0].CardInstanceID)
	beforeVersion := card.ZoneVersion
	targets[0].Controller = game.Player3
	targets[1].Counters.Add(counter.Shield, 1)
	ctx := &programContext{
		engine: engine, game: g, obj: obj, active: true, log: &TurnLog{},
		specs: []game.TargetSpec{{MinTargets: 0, MaxTargets: 3, Allow: game.TargetAllowPermanent}},
	}
	outcome, err := ctx.Destroy(game.Destroy{Object: game.AllTargetPermanentsReference(0)})
	if err != nil || !outcome.Available || len(outcome.Members) != 2 {
		t.Fatalf("batch = %+v, %v", outcome, err)
	}
	moved, protected := outcome.Members[0], outcome.Members[1]
	if moved.Before.ObjectID != targets[0].ObjectID || moved.Before.Controller != game.Player3 ||
		moved.Before.ZoneCards[0].ZoneVersion != beforeVersion ||
		!moved.Attempted || !moved.Moved || moved.Destination != zone.Graveyard ||
		len(moved.Arrivals) != 1 || moved.Arrivals[0].CardID != card.ID || moved.Arrivals[0].ZoneVersion != card.ZoneVersion ||
		card.ZoneVersion <= beforeVersion {
		t.Fatalf("departure lost its actual before/arrival correlation: %+v", moved)
	}
	if !protected.Attempted || protected.Moved || len(protected.Arrivals) != 0 ||
		protected.Before.Counters.Get(counter.Shield) != 1 || targets[1].Counters.Get(counter.Shield) != 0 {
		t.Fatal("protected member was omitted or its before snapshot mutated")
	}
	outcome.Members[0].Arrivals[0].ZoneVersion = 999
	outcome.Members[1].Before.Counters.Add(counter.Shield, 10)
	lki, _ := lastKnownObject(g, targets[0].ObjectID)
	if card.ZoneVersion == 999 || lki.ZoneCards[0].ZoneVersion == 999 || targets[1].Counters.Get(counter.Shield) != 0 {
		t.Fatal("public outcomes expose mutable engine containers")
	}
}

func TestAuthoredDestroyKnownEmptyVersusUnavailableAndDeduplication(t *testing.T) {
	t.Parallel()
	g, engine, targets := terastodonStack(t, 1)
	obj, _ := g.Stack.Peek()
	ctx := &programContext{
		engine: engine, game: g, obj: obj, active: true, log: &TurnLog{},
		specs: []game.TargetSpec{{MinTargets: 0, MaxTargets: 3, Allow: game.TargetAllowPermanent}},
	}
	obj.Targets = append(obj.Targets, obj.Targets[0])
	obj.TargetCounts = []int{2}
	outcome, err := ctx.Destroy(game.Destroy{Object: game.AllTargetPermanentsReference(0)})
	if err != nil || len(outcome.Members) != 1 || outcome.Members[0].Before.ObjectID != targets[0].ObjectID {
		t.Fatalf("duplicate original incarnations were processed twice: %+v, %v", outcome, err)
	}
	outcome, err = ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	if err != nil || outcome.Available || len(outcome.Members) != 0 {
		t.Fatal("an unavailable single reference became a known successful empty result")
	}
	obj.Targets, obj.TargetCounts = nil, []int{0}
	outcome, err = ctx.Destroy(game.Destroy{Object: game.AllTargetPermanentsReference(0)})
	if err != nil || !outcome.Available || len(outcome.Members) != 0 {
		t.Fatal("a declared zero-target batch lost its known availability")
	}
	ctx.active = false
	if _, err := ctx.Destroy(game.Destroy{Object: game.AllTargetPermanentsReference(0)}); err == nil {
		t.Fatal("expired context could destroy")
	}
	if _, err := ctx.CreateTokens(&game.CardDef{CardFace: game.CardFace{Name: "Token"}}, cardprogram.TokenRecipient{Player: game.Player1, Count: 1}); err == nil {
		t.Fatal("expired context could create tokens")
	}
}

type terastodonPolicy struct {
	simPassPolicy

	targets []game.Target
}

func (p terastodonPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			if len(option.Targets) == len(p.targets) && slices.Equal(option.Targets, p.targets) {
				return []int{option.Index}
			}
		}
		return nil
	}
	return request.DefaultSelection
}

func TestTerastodonRealEnterAnnouncementAndSimulatorClone(t *testing.T) {
	t.Parallel()
	for count := range 4 {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			t.Parallel()
			g, engine, targets := terastodonStack(t, count)
			old, _ := g.Stack.Pop()
			source, _ := g.GetCardInstance(old.SourceCardID)
			sourcePermanent := permanentByCardID(g, source.ID)
			removePermanentFromBattlefield(g, sourcePermanent.ObjectID)
			addImplementationSpellToStack(g, game.Player1, source.Def, nil)
			addCombatPermanent(g, game.Player2, vanillaCreature("Not a legal target", 2, 2))
			want := make([]game.Target, len(targets))
			for i, target := range targets {
				want[i] = game.PermanentTarget(target.ObjectID)
			}
			policies := simPassPolicies()
			policies[game.Player1] = terastodonPolicy{targets: want}
			setMainPhasePriority(g, game.Player1)
			simulator := engine.newSearchContext(g, game.Player1).Simulator()
			branch := simulator.ResolvePriority(g, policies)
			if !branch.Stack.IsEmpty() || namedTokenCount(branch, game.Player2, "Elephant") != count {
				t.Fatal("production ETB program did not resolve in the Simulator clone")
			}
			if namedTokenCount(g, game.Player2, "Elephant") != 0 || len(g.Battlefield) != count+1 {
				t.Fatal("Simulator mutated the original")
			}
			for _, target := range targets {
				if !branch.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
					t.Fatal("announced legal target was not destroyed")
				}
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, policies, &log)
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, policies, &log) {
				t.Fatal("real Terastodon entry did not announce a triggered ability")
			}
			trigger, _ := g.Stack.Peek()
			if !slices.Equal(trigger.Targets, want) {
				t.Fatal("trigger announcement did not select the requested legal target set")
			}
			for _, decision := range log.Choices {
				if decision.Request.Kind == game.ChoiceTarget {
					for _, option := range decision.Request.Options {
						for _, target := range option.Targets {
							permanent, _ := permanentByObjectID(g, target.PermanentID)
							if permanent != nil && permanentHasType(g, permanent, types.Creature) {
								t.Fatal("target announcement included a creature")
							}
						}
					}
				}
			}
			engine.resolveTopOfStackWithChoices(g, policies, &log)
			if namedTokenCount(g, game.Player2, "Elephant") != count {
				t.Fatal("scripted actual stack path diverged from Simulator")
			}
		})
	}
}
