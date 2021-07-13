package hash

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestHashMeerXKeccakV1(t *testing.T) {
	str := "helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel"
	h := HashMeerXKeccakV1([]byte(str))
	assert.Equal(t, hex.EncodeToString(h[:]), "046bb4ee0e487afb53c428f5d18f2875951f80330fe39870011ddac6a9c06b1c")
	str = "12000000c24900a189710157b8d020f2d09068ec4ee1eacae6b38753a6304894d0e2104b80eb912fbeb3b6735a4acfb89fd199e0f0abc50a48dccd52e2245e0619755468000000000000000000000000000000000000000000000000000000000000000068db061fda20d75f088f10000000000000"
	b, _ := hex.DecodeString(str)
	h = HashMeerXKeccakV1(b)
	assert.Equal(t, hex.EncodeToString(h[:]), "84499238afb6b544e01032f2ea73d5692ef5d29775daa3a778bfbd2efbb90000")
	str = "120000005b42f3a292059337e2097c4077d6578adf4253c63ab74a51d57a3cf330003267b012b719ebfb86d8ea7864307bd0b743980ead38d1f2c06919fc45d6eb604ed9000000000000000000000000000000000000000000000000000000000000000068db061fa421d75f086605000000000000"
	b, _ = hex.DecodeString(str)
	h = HashMeerXKeccakV1(b)
	assert.Equal(t, hex.EncodeToString(h[:]), "3b94fab4d25980aca4512b86b32bc96fd7cb8f8fdf3f265ffde45f72688e0100")

	str = "120000003d23a9a7eb43ac14e7a00fb85074633f4279c1748182582fe69843f945512392c04a59e4d676f01b9f7755fe033602e0529dea90df046ecd3515900aeff9eab90000000000000000000000000000000000000000000000000000000000000000ffff032057d8e85f08899d0de000000000"
	b, _ = hex.DecodeString(str)
	h = HashMeerXKeccakV1(b)
	fmt.Println(hex.EncodeToString(h[:]))
}
