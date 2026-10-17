package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestETBSuppressionCapturedDuringAuthoredResolution(t *testing.T) {
	for _, destroyAfter := range []bool{true, false} {
		for _, doubled := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			nornDef := cards.NewDefaultRegistry().Lookup("Elesh Norn, Mother of Machines")
			var targets []game.Target
			if destroyAfter {
				norn := addCombatPermanent(g, game.Player2, nornDef)
				targets = []game.Target{game.PermanentTarget(norn.ObjectID)}
			}
			if doubled {
				addCombatPermanent(g, game.Player1, enteringTriggerDoublerDef(nil))
			}
			token := selfEntersTriggerSourceDef([]types.Card{types.Creature})
			token.Power, token.Toughness = opt.Val(game.PT{Value: 2}), opt.Val(game.PT{Value: 2})
			ref := game.ProgramReference{ID: "test/etb-suppression", Version: "1"}
			catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
				Reference: ref,
				Resolve: func(ctx cardprogram.Context) error {
					if _, err := ctx.CreateTokens(token, cardprogram.TokenRecipient{Player: game.Player1, Count: 1}); err != nil {
						return err
					}
					if destroyAfter {
						_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
						return err
					}
					_, err := ctx.CreateTokens(nornDef, cardprogram.TokenRecipient{Player: game.Player2, Count: 1})
					return err
				},
			}})
			if err != nil {
				t.Fatal(err)
			}
			e := NewEngineWithPrograms(nil, catalog)
			mode := game.Mode{Program: ref}
			if destroyAfter {
				mode.Targets = []game.TargetSpec{{Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1}}
			}
			addImplementationSpellToStack(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Token then change Norn", Types: []types.Card{types.Sorcery}, SpellAbility: opt.Val(mode.Ability()),
			}}, targets)
			before := g.Clone()
			e.resolveTopOfStack(g, &TurnLog{})
			if destroyAfter && g.Players[game.Player2].Graveyard.Size() != 1 {
				t.Fatal("authored program did not destroy Norn after the token entered")
			}
			want := 0
			if !destroyAfter {
				want = 1
				if doubled {
					want = 2
				}
			}
			for _, result := range []*game.Game{g, g.Clone()} {
				e.putTriggeredAbilitiesOnStack(result)
				if result.Stack.Size() != want {
					t.Fatalf("destroyAfter=%v doubled=%v: %d entry triggers, want %d", destroyAfter, doubled, result.Stack.Size(), want)
				}
			}
			branch := e.Simulator().ResolvePriority(before, simPassPolicies())
			if branch.Players[game.Player1].Life != 40+want || before.Players[game.Player1].Life != 40 ||
				before.Stack.Size() != 1 {
				t.Fatal("Simulator changed suppression, multipliers, or the original branch")
			}
		}
	}
}

func TestETBSuppressionIgnoresLaterSuppressorAndSourceChanges(t *testing.T) {
	for _, change := range []string{"leave", "phase out", "lose abilities", "controller", "source controller"} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngine(nil)
		norn := addCombatPermanent(g, game.Player2, opponentEnteringSuppressorDef())
		source, ok := createTokenPermanentWithChoices(e, g, game.Player1,
			selfEntersTriggerSourceDef([]types.Card{types.Creature}), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !ok {
			t.Fatal("could not enter trigger source")
		}
		switch change {
		case "leave":
			movePermanentToZone(g, norn, zone.Graveyard)
		case "phase out":
			norn.PhasedOut = true
		case "lose abilities":
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), Layer: game.LayerAbility, AffectedObjectID: norn.ObjectID, RemoveAllAbilities: true,
			})
		case "controller":
			norn.Controller = game.Player1
		case "source controller":
			source.Controller = game.Player2
		default:
			t.Fatalf("unknown change %q", change)
		}
		if e.putTriggeredAbilitiesOnStack(g) {
			t.Fatalf("later %s resurrected a prevented trigger", change)
		}
	}
}

func TestETBSuppressionUsesSimultaneousPostEventState(t *testing.T) {
	for _, simultaneous := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			sourceID := addCardToLibrary(g, game.Player1, selfEntersTriggerSourceDef([]types.Card{types.Creature}))
			nornID := addCardToLibrary(g, game.Player2, opponentEnteringSuppressorDef())
			resolved := []resolvedBattlefieldCard{
				{card: g.CardInstances[sourceID], fromZone: zone.Library, controller: game.Player1},
				{card: g.CardInstances[nornID], fromZone: zone.Library, controller: game.Player2},
			}
			r := effectResolver{engine: e, game: g, obj: &game.StackObject{Controller: game.Player1}, log: &TurnLog{}}
			if simultaneous {
				if reverse {
					slices.Reverse(resolved)
				}
				r.putResolvedCardsOnBattlefieldCollecting(resolved, nil, permanentCreationOptions{})
			} else {
				for _, card := range resolved {
					r.putResolvedCardOnBattlefieldValue(card.card, card.fromZone, card.controller, nil, permanentCreationOptions{})
				}
			}
			e.putTriggeredAbilitiesOnStack(g)
			want := 1
			if simultaneous {
				want = 0
			}
			if g.Stack.Size() != want {
				t.Fatalf("simultaneous=%v reverse=%v: %d triggers, want %d", simultaneous, reverse, g.Stack.Size(), want)
			}
		}
	}
}

func TestETBSuppressionSyntheticCompatibilityStillFilters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addCombatPermanent(g, game.Player2, opponentEnteringSuppressorDef())
	source := addCombatPermanent(g, game.Player1, selfEntersTriggerSourceDef([]types.Card{types.Creature}))
	g.AppendEvent(game.Event{
		Kind: game.EventZoneChanged, PermanentID: source.ObjectID, CardID: source.CardInstanceID,
		Controller: game.Player1, ToZone: zone.Battlefield,
	})
	if e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("synthetic uncaptured event bypassed suppression")
	}
}

func TestETBSuppressionWarpWorldStageBoundaries(t *testing.T) {
	for _, laterStage := range []bool{false, true} {
		suppressor := opponentEnteringSuppressorDef()
		if laterStage {
			suppressor.Types = []types.Card{types.Enchantment}
		}
		g, e, _ := warpWorldLibraryFixture(t, [game.NumPlayers][]*game.CardDef{
			game.Player1: {selfEntersTriggerSourceDef([]types.Card{types.Creature})},
			game.Player2: {suppressor},
		})
		e.resolveTopOfStack(g, &TurnLog{})
		e.putTriggeredAbilitiesOnStack(g)
		want := 0
		if laterStage {
			want = 1
		}
		if g.Stack.Size() != want {
			t.Fatalf("later-stage suppressor=%v: %d triggers, want %d", laterStage, g.Stack.Size(), want)
		}
	}
}
