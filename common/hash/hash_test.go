package hash

import (
	"encoding/hex"
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

func TestHashMeerCrypto(t *testing.T) {
	str := "helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel"
	h := HashMeerCrypto([]byte(str))
	assert.Equal(t, hex.EncodeToString(h[:]), "250bcbab5c6959fd0325c513ff1e95e05ef1bbbd39214a050b44e52046573b5a")
	str = "120000009cbbf9987538e02437af518ee0e262e1381e6cd075b579ea08a2ba95242d5dfe740a01155aa4ae1fec77c127d19d0c93ccf745d879af54bcb0c4724d04d4570e0000000000000000000000000000000000000000000000000000000000000000ffff03209638d65f08e600000000000000"
	b, _ := hex.DecodeString(str)
	h = HashMeerCrypto(b)
	assert.Equal(t, hex.EncodeToString(h[:]), "fbde247d9418ad7886572e2e90c76ed05ac58b0674088ca2b9596e9479af1003")
	str = "12000000352e2843a0536e80a2af770b87be244ea0d007f368e57c05dcbf58c1d98f3a95b47b480124466fe5088a3490a8729e3f0c57a85204ad2ab5ab502a6a6d46a69a0000000000000000000000000000000000000000000000000000000000000000d0b3061f283ad65f08f12b000000000000"
	b, _ = hex.DecodeString(str)
	h = HashMeerCrypto(b)
	assert.Equal(t, hex.EncodeToString(h[:]), "5274d9733b48a7513a57950c59ee897106bdcc8313ea60b777b1c3ad739d0500")
}
