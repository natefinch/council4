package mana

// Unit is a spendable unit of mana. Color records the mana color or colorless;
// Snow records whether the mana was produced by a snow source for costs such as
// {S}; FromCreature records whether the mana was produced by a creature source
// so cards such as Inga and Esika can count "mana from creatures" spent to cast
// a spell. Provenance is fixed when the mana is produced, so it survives the
// source later changing type or leaving the battlefield.
type Unit struct {
	Color        Color
	Snow         bool
	FromCreature bool
}
