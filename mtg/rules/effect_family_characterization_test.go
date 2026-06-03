package rules

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestHighFrequencyEffectFamilyCharacterizationSnapshot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	searched := addCardToLibrary(g, game.Player1, phaseZeroCard("Searched Creature", types.Creature, 0))
	_ = addCardToLibrary(g, game.Player1, phaseZeroCard("Unmatched Instant", types.Instant, 0))
	_ = addCardToLibrary(g, game.Player1, phaseZeroCard("Drawn Card", types.Sorcery, 0))
	artifact := addCombatPermanent(g, game.Player1, phaseZeroCard("Dormant Relic", types.Artifact, 0))
	firstFighter := addCombatPermanent(g, game.Player1, phaseZeroCard("Valiant Cub", types.Creature, 2))
	secondFighter := addCombatPermanent(g, game.Player2, phaseZeroCard("Hill Giant", types.Creature, 3))
	four := game.PT{Value: 4}
	sourceID := addLinkedResultSpellToStackForController(g, game.Player1, []game.Effect{
		{
			Type:   game.EffectChoose,
			LinkID: "chosen-color",
			Choice: opt.Val(game.ResolutionChoice{
				Kind: game.ResolutionChoiceMana,
			}),
		},
		{Type: game.EffectAddMana, Amount: 2, ChoiceLinkID: "chosen-color"},
		{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController},
		{Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
		{
			Type:        game.EffectSearch,
			Amount:      1,
			TargetIndex: game.TargetIndexController,
			Search: opt.Val(game.SearchSpec{
				SourceZone:  game.ZoneLibrary,
				Destination: game.ZoneHand,
				CardType:    opt.Val(types.Creature),
				Reveal:      true,
			}),
		},
		{
			Type:        game.EffectApplyContinuous,
			TargetIndex: 1,
			ContinuousEffects: []game.ContinuousEffect{
				{Layer: game.LayerType, AddTypes: []types.Card{types.Creature}},
				{Layer: game.LayerPowerToughnessSet, SetPower: opt.Val(four), SetToughness: opt.Val(four)},
			},
		},
		{
			Type:           game.EffectModifyPT,
			TargetIndex:    2,
			PowerDelta:     1,
			ToughnessDelta: 1,
			UntilEndOfTurn: true,
		},
		{
			Type:               game.EffectFight,
			TargetIndex:        2,
			RelatedTargetIndex: opt.Val(3),
		},
	}, []game.Target{
		game.PlayerTarget(game.Player2),
		game.PermanentTarget(artifact.ObjectID),
		game.PermanentTarget(firstFighter.ObjectID),
		game.PermanentTarget(secondFighter.ObjectID),
	})
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{3}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	labels := map[id.ID]string{
		artifact.ObjectID:      "Dormant Relic",
		firstFighter.ObjectID:  "Valiant Cub",
		secondFighter.ObjectID: "Hill Giant",
	}
	artifactToughness, artifactHasToughness := effectiveToughness(g, artifact)
	snapshot := []string{
		fmt.Sprintf("resolve: %s", log.Resolves[0].Result),
		fmt.Sprintf("choices: %s", phaseZeroChoiceSnapshot(&log)),
		fmt.Sprintf("life: P1=%d P2=%d", g.Players[game.Player1].Life, g.Players[game.Player2].Life),
		fmt.Sprintf("mana: total=%d red=%d", g.Players[game.Player1].ManaPool.Total(), g.Players[game.Player1].ManaPool.Amount(mana.R)),
		fmt.Sprintf("hand: %s", cardList(g, g.Players[game.Player1].Hand.All())),
		fmt.Sprintf("library: %s", cardList(g, g.Players[game.Player1].Library.All())),
		fmt.Sprintf("searched: in-hand=%t", g.Players[game.Player1].Hand.Contains(searched)),
		fmt.Sprintf("artifact: creature=%t power=%d toughness=%d/%t", permanentHasType(g, artifact, types.Creature), effectivePower(g, artifact), artifactToughness, artifactHasToughness),
		fmt.Sprintf("fighters: cub-damage=%d giant-damage=%d", firstFighter.MarkedDamage, secondFighter.MarkedDamage),
		fmt.Sprintf("continuous-effects: %d", len(g.ContinuousEffects)),
		fmt.Sprintf("deaths: %s", phaseZeroDeathSnapshot(labels, deaths)),
		"events:",
	}
	snapshot = append(snapshot, phaseZeroEventSnapshot(g, labels)...)
	want := []string{
		"resolve: graveyard",
		"choices: resolution selected=[3] fallback=false",
		"life: P1=40 P2=37",
		"mana: total=2 red=2",
		"hand: [Searched Creature, Drawn Card]",
		"library: [Unmatched Instant]",
		"searched: in-hand=true",
		"artifact: creature=true power=4 toughness=4/true",
		"fighters: cub-damage=3 giant-damage=3",
		"continuous-effects: 3",
		"deaths: [Valiant Cub:lethal damage, Hill Giant:lethal damage]",
		"events:",
		"zone Drawn Card Library->Hand",
		"draw P1 Drawn Card",
		"life-lost P2 amount=3",
		fmt.Sprintf("damage player P2 amount=3 source=%s", cardName(g, sourceID)),
		"reveal P1 Searched Creature from Library",
		"zone Searched Creature Library->Hand",
		"fight Valiant Cub->Hill Giant",
		"fight Hill Giant->Valiant Cub",
		"damage permanent Hill Giant amount=3 source=Valiant Cub",
		"damage permanent Valiant Cub amount=3 source=Hill Giant",
		fmt.Sprintf("zone %s Stack->Graveyard", cardName(g, sourceID)),
		fmt.Sprintf("spell-resolved %s", cardName(g, sourceID)),
		"zone Valiant Cub Battlefield->Graveyard",
		"died Valiant Cub",
		"zone Hill Giant Battlefield->Graveyard",
		"died Hill Giant",
	}
	if got, want := strings.Join(snapshot, "\n"), strings.Join(want, "\n"); got != want {
		t.Fatalf("snapshot:\n%s\nwant:\n%s", got, want)
	}
}

func phaseZeroCard(name string, cardType types.Card, power int) *game.CardDef {
	face := game.CardFace{Name: name, Types: []types.Card{cardType}}
	if cardType == types.Creature && power > 0 {
		pt := game.PT{Value: power}
		face.Power = opt.Val(pt)
		face.Toughness = opt.Val(pt)
	}
	return &game.CardDef{CardFace: face}
}

func phaseZeroChoiceSnapshot(log *TurnLog) string {
	if len(log.Choices) == 0 {
		return "[]"
	}
	choice := log.Choices[0]
	return fmt.Sprintf("%s selected=%v fallback=%t", phaseZeroChoiceKind(choice.Request.Kind), choice.Selected, choice.UsedFallback)
}

func phaseZeroChoiceKind(kind game.ChoiceKind) string {
	if kind == game.ChoiceResolution {
		return "resolution"
	}
	return fmt.Sprintf("choice-%d", kind)
}

func phaseZeroDeathSnapshot(labels map[id.ID]string, deaths []PermanentDeathLog) string {
	if len(deaths) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(deaths))
	for _, death := range deaths {
		parts = append(parts, fmt.Sprintf("%s:%s", labels[death.Permanent], death.Reason))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func phaseZeroEventSnapshot(g *game.Game, labels map[id.ID]string) []string {
	snapshot := make([]string, 0, len(g.Events))
	for _, event := range g.Events {
		snapshot = append(snapshot, phaseZeroEventLine(g, labels, event))
	}
	return snapshot
}

func phaseZeroEventLine(g *game.Game, labels map[id.ID]string, event game.GameEvent) string {
	switch event.Kind {
	case game.EventLifeLost:
		return fmt.Sprintf("life-lost %s amount=%d", phaseZeroPlayer(event.Player), event.Amount)
	case game.EventDamageDealt:
		return phaseZeroDamageEventLine(g, labels, event)
	case game.EventZoneChanged:
		return fmt.Sprintf("zone %s %s->%s", phaseZeroEventObjectName(g, labels, event), event.FromZone, event.ToZone)
	case game.EventCardDrawn:
		return fmt.Sprintf("draw %s %s", phaseZeroPlayer(event.Player), cardName(g, event.CardID))
	case game.EventCardRevealed:
		return fmt.Sprintf("reveal %s %s from %s", phaseZeroPlayer(event.Player), cardName(g, event.CardID), event.FromZone)
	case game.EventFight:
		return fmt.Sprintf("fight %s->%s", labels[event.PermanentID], labels[event.RelatedPermanentID])
	case game.EventSpellResolved:
		return fmt.Sprintf("spell-resolved %s", cardName(g, event.CardID))
	case game.EventPermanentDied:
		return fmt.Sprintf("died %s", labels[event.PermanentID])
	default:
		return fmt.Sprintf("event-%d", event.Kind)
	}
}

func phaseZeroDamageEventLine(g *game.Game, labels map[id.ID]string, event game.GameEvent) string {
	if event.DamageRecipient == game.DamageRecipientPlayer {
		return fmt.Sprintf("damage player %s amount=%d source=%s", phaseZeroPlayer(event.Player), event.Amount, cardName(g, event.SourceID))
	}
	return fmt.Sprintf("damage permanent %s amount=%d source=%s", labels[event.PermanentID], event.Amount, cardName(g, event.SourceID))
}

func phaseZeroEventObjectName(g *game.Game, labels map[id.ID]string, event game.GameEvent) string {
	if event.PermanentID != 0 {
		if label := labels[event.PermanentID]; label != "" {
			return label
		}
	}
	return cardName(g, event.CardID)
}

func phaseZeroPlayer(player game.PlayerID) string {
	return fmt.Sprintf("P%d", int(player)+1)
}
