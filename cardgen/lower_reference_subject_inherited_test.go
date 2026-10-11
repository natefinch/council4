package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func inheritedOriginalSourceCards() []ScryfallCard {
	return []ScryfallCard{
		{Name: "Crovax the Cursed", Layout: "normal", ManaCost: "{2}{B}{B}", Colors: []string{"B"},
			TypeLine: "Legendary Creature \u2014 Vampire Noble", Power: new("0"), Toughness: new("0"),
			OracleText: "Crovax enters with four +1/+1 counters on it.\nAt the beginning of your upkeep, you may sacrifice a creature. If you do, put a +1/+1 counter on Crovax. If you don't, remove a +1/+1 counter from Crovax.\n{B}: Crovax gains flying until end of turn."},
		{Name: "Dire Fleet Warmonger", Layout: "normal", ManaCost: "{1}{B}{R}", Colors: []string{"B", "R"},
			TypeLine: "Creature \u2014 Orc Pirate", Power: new("3"), Toughness: new("3"),
			OracleText: "At the beginning of combat on your turn, you may sacrifice another creature. If you do, this creature gets +2/+2 and gains trample until end of turn. (It can deal excess combat damage to the player or planeswalker it's attacking.)"},
		{Name: "Kill-Zone Acrobat", Layout: "normal", ManaCost: "{2}{B}", Colors: []string{"B"},
			TypeLine: "Creature \u2014 Human Soldier", Power: new("3"), Toughness: new("2"),
			OracleText: "Whenever this creature attacks, you may sacrifice another creature or artifact. If you do, this creature gains flying until end of turn."},
		{Name: "Mukotai Soulripper", Layout: "normal", ManaCost: "{1}{B}", Colors: []string{"B"},
			TypeLine: "Artifact \u2014 Vehicle", Power: new("4"), Toughness: new("3"),
			OracleText: "Whenever this Vehicle attacks, you may sacrifice another artifact or creature. If you do, put a +1/+1 counter on this Vehicle and it gains menace until end of turn.\nCrew 2 (Tap any number of creatures you control with total power 2 or more: This Vehicle becomes an artifact creature until end of turn.)"},
		{Name: "Ragefire Hellkite", Layout: "normal", ManaCost: "{4}{R}{R}", Colors: []string{"R"},
			TypeLine: "Creature \u2014 Dragon", Power: new("5"), Toughness: new("3"),
			OracleText: "Flying\nWhenever this creature attacks, you may sacrifice another creature. If you do, this creature gains double strike until end of turn."},
		{Name: "Angelic Renewal", Layout: "normal", ManaCost: "{1}{W}", Colors: []string{"W"}, TypeLine: "Enchantment",
			OracleText: "Whenever a creature is put into your graveyard from the battlefield, you may sacrifice this enchantment. If you do, return that card to the battlefield."},
		{Name: "Dreamcatcher", Layout: "normal", ManaCost: "{U}", Colors: []string{"U"},
			TypeLine: "Creature \u2014 Spirit", Power: new("1"), Toughness: new("1"),
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may sacrifice this creature. If you do, draw a card."},
		{Name: "Eden, Seat of the Sanctum", Layout: "normal", TypeLine: "Land \u2014 Town",
			OracleText: "{T}: Add {C}.\n{5}, {T}: Mill two cards. Then you may sacrifice this land. When you do, return another target permanent card from your graveyard to your hand."},
		{Name: "Flamespeaker's Will", Layout: "normal", ManaCost: "{R}", Colors: []string{"R"}, TypeLine: "Enchantment \u2014 Aura",
			OracleText: "Enchant creature you control\nEnchanted creature gets +1/+1.\nWhenever enchanted creature deals combat damage to a player, you may sacrifice this Aura. If you do, destroy target artifact."},
		{Name: "Grave Peril", Layout: "normal", ManaCost: "{1}{B}", Colors: []string{"B"}, TypeLine: "Enchantment",
			OracleText: "When a nonblack creature enters, sacrifice this enchantment. If you do, destroy that creature."},
		{Name: "Marit Lage's Slumber", Layout: "normal", ManaCost: "{1}{U}", Colors: []string{"U"}, TypeLine: "Legendary Snow Enchantment",
			OracleText: "Whenever Marit Lage's Slumber or another snow permanent you control enters, scry 1.\nAt the beginning of your upkeep, if you control ten or more snow permanents, sacrifice Marit Lage's Slumber. If you do, create Marit Lage, a legendary 20/20 black Avatar creature token with flying and indestructible."},
		{Name: "Mortal Obstinacy", Layout: "normal", ManaCost: "{W}", Colors: []string{"W"}, TypeLine: "Enchantment \u2014 Aura",
			OracleText: "Enchant creature you control\nEnchanted creature gets +1/+1.\nWhenever enchanted creature deals combat damage to a player, you may sacrifice this Aura. If you do, destroy target enchantment."},
		{Name: "Mystery Key", Layout: "normal", ManaCost: "{1}{U}", Colors: []string{"U"}, TypeLine: "Artifact \u2014 Equipment",
			OracleText: "When equipped creature deals combat damage to a player, sacrifice this Equipment. If you do, draw three cards.\nEquip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)"},
		{Name: "Promise of Bunrei", Layout: "normal", ManaCost: "{2}{W}", Colors: []string{"W"}, TypeLine: "Enchantment",
			OracleText: "When a creature you control dies, sacrifice this enchantment. If you do, create four 1/1 colorless Spirit creature tokens."},
		{Name: "Smoke Bomb", Layout: "normal", ManaCost: "{3}", TypeLine: "Artifact",
			OracleText: "Flash\nAll creatures have shroud. (They can't be the targets of spells or abilities.)\nAt the beginning of your upkeep, sacrifice this artifact. When you do, target creature you control can't be blocked this turn."},
		{Name: "Throwing Knife", Layout: "normal", ManaCost: "{2}", TypeLine: "Artifact \u2014 Equipment",
			OracleText: "Equipped creature gets +2/+0.\nWhenever equipped creature attacks, you may sacrifice this Equipment. If you do, this Equipment deals 2 damage to any target.\nEquip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)"},
	}
}

func visitReferenceSubjectContent(content game.AbilityContent, visit func(game.Primitive)) {
	for _, mode := range content.Modes {
		for _, instruction := range mode.Sequence {
			visit(instruction.Primitive)
			if instruction.Primitive.Kind() == game.PrimitiveCreateReflexiveTrigger {
				trigger, ok := instruction.Primitive.(game.CreateReflexiveTrigger)
				if !ok {
					panic("CreateReflexiveTrigger kind has an incompatible primitive")
				}
				visitReferenceSubjectContent(trigger.Trigger.Content, visit)
			}
		}
	}
}

func TestReferenceSubjectAllInheritedOriginalSourceRoutes(t *testing.T) {
	for index, card := range inheritedOriginalSourceCards() {
		t.Run(card.Name, func(t *testing.T) {
			face := lowerSingleFace(t, &card)
			found := 0
			visit := func(primitive game.Primitive) {
				if index < 5 && primitive.Kind() == game.PrimitiveApplyContinuous {
					grant, ok := primitive.(game.ApplyContinuous)
					if !ok {
						t.Fatalf("unexpected continuous primitive %T", primitive)
					}
					if grant.Object.Exists {
						if grant.Object.Val != game.SourcePermanentReference() {
							t.Fatalf("self-keyword subject=%v, want original permanent", grant.Object.Val)
						}
						found++
					}
				}
				if index >= 5 && primitive.Kind() == game.PrimitiveSacrifice {
					sacrifice, ok := primitive.(game.Sacrifice)
					if !ok {
						t.Fatalf("unexpected sacrifice primitive %T", primitive)
					}
					if sacrifice.Object != game.SourcePermanentReference() {
						t.Fatalf("self-sacrifice subject=%v, want original permanent", sacrifice.Object)
					}
					found++
				}
			}
			for _, ability := range face.ActivatedAbilities {
				visitReferenceSubjectContent(ability.Content, visit)
			}
			for _, ability := range face.TriggeredAbilities {
				visitReferenceSubjectContent(ability.Content, visit)
			}
			if found == 0 {
				t.Fatal("complete original lost its audited self subject")
			}
		})
	}
}
