package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestRendererPreservesLinkedProducerContract(t *testing.T) {
	const producer game.LinkedProducerKey = "test/exile"
	body := game.Mode{Sequence: []game.Instruction{{Primitive: game.MovePermanent{
		Object: game.SourcePermanentReference(), Destination: zone.Exile, PublishLinkedProducer: producer,
	}}}}.Ability()
	body.LinkedProducers = []game.LinkedProducerKey{producer}
	def := &game.CardDef{CardFace: game.CardFace{Name: "Producer", Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{Content: body}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{Event: game.EventPermanentDied, Source: game.TriggerSourceSelf}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
				Source: game.OriginalSourceLinkedBattlefieldSource(producer),
			}}}}.Ability(),
		}},
	}}
	source, err := (Renderer{}).RenderCardSource(&ScryfallCard{Name: "Producer", Layout: "normal"},
		[]*game.CardDef{def}, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	source = strings.Join(strings.Fields(source), " ")
	for _, fragment := range []string{
		`LinkedProducers: []game.LinkedProducerKey{game.LinkedProducerKey("test/exile")}`,
		`PublishLinkedProducer: game.LinkedProducerKey("test/exile")`,
		`game.OriginalSourceLinkedBattlefieldSource(game.LinkedProducerKey("test/exile"))`,
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("linked binding %q lost:\n%s", fragment, source)
		}
	}
	move, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.MoveCard{
		Card: game.CardReference{Kind: game.CardReferenceEvent}, FromZone: zone.Graveyard,
		Destination: zone.Exile, PublishLinkedProducer: producer,
	})
	if err != nil || !strings.Contains(move, `PublishLinkedProducer: game.LinkedProducerKey("test/exile")`) {
		t.Fatalf("MoveCard producer rendering=%s, %v", move, err)
	}
}

func TestLegacyLinkedDeathReturnsRemainFailClosed(t *testing.T) {
	three, four, five := "3", "4", "5"
	for _, card := range []*ScryfallCard{
		{Name: "Helvault", TypeLine: "Legendary Artifact", ManaCost: "{3}", Layout: "normal",
			OracleText: "{1}, {T}: Exile target creature you control.\n{7}, {T}: Exile target creature you don't control.\nWhen Helvault is put into a graveyard from the battlefield, return all cards exiled with it to the battlefield under their owners' control."},
		{Name: "Trove Warden", TypeLine: "Creature — Cat Beast", ManaCost: "{2}{W}{W}", Layout: "normal", Power: &three, Toughness: &four,
			OracleText: "Vigilance\nLandfall — Whenever a land you control enters, exile target permanent card with mana value 3 or less from your graveyard.\nWhen this creature dies, put each permanent card exiled with it onto the battlefield under the control of that card's owner."},
		{Name: "Underworld Sentinel", TypeLine: "Creature — Human Soldier", ManaCost: "{3}{B}{B}", Layout: "normal", Power: &four, Toughness: &five,
			OracleText: "Whenever this creature attacks, exile target creature card from your graveyard.\nWhen this creature dies, put all cards exiled with it onto the battlefield."},
	} {
		t.Run(card.Name, func(t *testing.T) {
			source, diagnostics, err := GenerateExecutableCardSource(card, "test")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatal("historical incorrect linked return no longer fails closed")
			}
		})
	}
}
