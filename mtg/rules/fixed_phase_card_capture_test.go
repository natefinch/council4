package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestFixedPhaseTargetCardProductDoesNotReuseFailedMove(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "Exile target creature card from your graveyard. You gain 1 life. Return that card to the battlefield at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	subject := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{
		Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}, TargetCounts: []int{1},
	}
	resolveInstruction(engine, g, obj, game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Graveyard}, nil)
	card, _ := g.GetCardInstance(subject.CardInstanceID)
	obj.Targets[0] = game.Target{
		Kind: game.TargetCard, CardID: subject.CardInstanceID,
		CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true,
	}
	for range 2 {
		engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	if len(g.DelayedTriggers) != 2 || g.DelayedTriggers[0].CapturedCardID != subject.CardInstanceID ||
		g.DelayedTriggers[1].CapturedCardID != 0 {
		t.Fatal("failed card exile reused an earlier actual card publication")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	returned, ok := findPermanentByCardID(g, subject.CardInstanceID)
	if !ok || returned.ObjectID == subject.ObjectID || returned.Controller != game.Player1 {
		t.Fatal("captured card did not return through the modeled battlefield primitive")
	}
}

func TestFixedPhaseSourceCardUsesDepartedIncarnation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
	handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Graveyard})
	def := &game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep, CapturedCard: opt.Val(game.SourcePermanentReference()),
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
			Card: game.CapturedCardReference(), FromZone: zone.Graveyard, Destination: zone.Hand,
		}}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, obj, def) || g.DelayedTriggers[0].CapturedCardID != source.CardInstanceID {
		t.Fatal("source card did not capture its actual reached graveyard incarnation")
	}
	handlePutOnBattlefield(r, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource})})
	returned, _ := findPermanentByCardID(g, source.CardInstanceID)
	newObj := *obj
	newObj.SourceID = returned.ObjectID
	newResolver := &effectResolver{engine: engine, game: g, obj: &newObj, log: &TurnLog{}}
	handleMovePermanent(newResolver, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Graveyard})
	if !scheduleDelayedTrigger(g, obj, def) || g.DelayedTriggers[1].CapturedCardID != 0 {
		t.Fatal("an old source reference captured a later source incarnation")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player2].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("an earlier delayed source-card action followed graveyard reentry")
	}
}

func TestFixedPhaseSourceAlreadyInGraveyardRequiresExactIncarnation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
	handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Graveyard})
	card, _ := g.GetCardInstance(source.CardInstanceID)
	obj.SourceZone = zone.Graveyard
	obj.SourceZoneVersion = card.ZoneVersion
	def := &game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep, CapturedCard: opt.Val(game.SourceCardPermanentReference()),
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(game.CapturedCardReference()),
		}}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, obj, def) || g.DelayedTriggers[0].CapturedCardID != card.ID {
		t.Fatal("the actual graveyard source incarnation was not captured")
	}
	stale := *obj
	stale.SourceZoneVersion--
	if !scheduleDelayedTrigger(g, &stale, def) || g.DelayedTriggers[1].CapturedCardID != 0 {
		t.Fatal("an unavailable graveyard source was replaced by current card identity")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	returned, ok := findPermanentByCardID(g, card.ID)
	if !ok || returned.Controller != game.Player2 {
		t.Fatal("the actual captured graveyard source did not return under its owner's control")
	}
}
