package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerBaelothBarritylEntertainer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Baeloth Barrityl, Entertainer",
		Layout:   "normal",
		ManaCost: "{4}{R}",
		TypeLine: "Legendary Creature — Elf Shaman",
		OracleText: "Creatures your opponents control with power less than Baeloth Barrityl's power are goaded. (They attack each combat if able and attack a player other than you if able.)\n" +
			"Whenever a goaded attacking or blocking creature dies, you create a Treasure token.\n" +
			"Choose a Background (You can have a Background as a second commander.)",
		Power:     new("2"),
		Toughness: new("5"),
	})

	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %#v, want goad and Choose a Background", face.StaticAbilities)
	}

	goad := face.StaticAbilities[0].Body.RuleEffects
	if len(goad) != 1 ||
		goad[0].Kind != game.RuleEffectGoaded ||
		goad[0].AffectedController != game.ControllerAny ||
		len(goad[0].PermanentTypes) != 1 ||
		goad[0].PermanentTypes[0] != types.Creature ||
		goad[0].AffectedSelection.Controller != game.ControllerOpponent ||
		!goad[0].AffectedSelection.PowerLessThanSource {
		t.Fatalf("goad rule effects = %#v", goad)
	}
	if len(face.StaticAbilities[1].Body.KeywordAbilities) != 1 {
		t.Fatalf("background ability = %#v", face.StaticAbilities[1])
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventPermanentDied ||
		!trigger.Trigger.Pattern.SubjectSelection.MatchGoaded ||
		trigger.Trigger.Pattern.SubjectSelection.CombatState != game.CombatStateAttackingOrBlocking {
		t.Fatalf("trigger pattern = %#v", trigger.Trigger.Pattern)
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("trigger sequence = %#v", sequence)
	}
	create, ok := sequence[0].Primitive.(game.CreateToken)
	token, tokenOK := create.Source.TokenDefRef()
	if !ok || create.Amount.Value() != 1 || !tokenOK || token.Name != string(types.Treasure) {
		t.Fatalf("create token = %#v", sequence[0].Primitive)
	}
}

func TestLowerGoadedAttackerControllerCreatesTreasure(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Goaded Attack Treasure",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a goaded creature attacks, its controller creates a Treasure token.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventAttackerDeclared ||
		!trigger.Trigger.Pattern.SubjectSelection.MatchGoaded {
		t.Fatalf("trigger pattern = %#v", trigger.Trigger.Pattern)
	}
	create, ok := trigger.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", trigger.Content.Modes[0].Sequence[0].Primitive)
	}
	if !create.Recipient.Exists ||
		create.Recipient.Val != game.ObjectControllerReference(game.EventPermanentReference()) {
		t.Fatalf("recipient = %#v, want event permanent controller", create.Recipient)
	}
}
