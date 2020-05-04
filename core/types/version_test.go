package types

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	versionData  = []byte{18, 0, 0, 0}   // uint32 => 18
	version1Data = []byte{18, 0, 17, 28} // uint32 => 470876178
)

func TestVersion(t *testing.T) {
	version := binary.LittleEndian.Uint32(versionData)
	version1 := binary.LittleEndian.Uint32(version1Data)
	assert.NotEqual(t, version1, version)
	assert.Equal(t, version1Data[:2], versionData[:2])
	header := &BlockHeader{
		Version: version1,
	}
	assert.Equal(t, version, header.GetVersion())
}
