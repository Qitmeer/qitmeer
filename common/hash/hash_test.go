package hash

type _Golden struct {
	out uint32
	in  string
}

var goldenTest = []_Golden{
	{0x00000000, ""},
	{0x00000000, "a"},
	{0x00000000, "ab"},
	{0x00000000, "abc"},
	{0x00000000, "abcd"},
	{0x00000000, "abcde"},
	{0x00000000, "abcdef"},
	{0x00000000, "abcdefg"},
	{0x00000000, "abcdefgh"},
	{0x00000000, "abcdefghi"},
	{0x00000000, "abcdefghij"},
	{0x00000000, "Discard medicine more than two years old."},
	{0x00000000, "He who has a shady past knows that nice guys finish last."},
	{0x00000000, "I wouldn't marry him with a ten foot pole."},
	{0x00000000, "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{0x00000000, "The days of the digital watch are numbered.  -Tom Stoppard"},
	{0x00000000, "Nepal premier won't resign."},
	{0x00000000, "For every action there is an equal and opposite government program."},
	{0x00000000, "His money is twice tainted: 'taint yours and 'taint mine."},
	{0x00000000, "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{0x00000000, "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{0x00000000, "size:  a.out:  bad magic"},
	{0x00000000, "The major problem is with sendmail.  -Mark Horton"},
	{0x00000000, "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{0x00000000, "If the enemy is within range, then so are you."},
	{0x00000000, "It's well we cannot hear the screams/That we create in others' dreams."},
	{0x00000000, "You remind me of a TV show, but that's all right: I watch it anyway."},
	{0x00000000, "C is as portable as Stonehedge!!"},
	{0x00000000, "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{0x00000000, "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{0x00000000, "How can you write a big system without C++?  -Paul Glick"},
}
