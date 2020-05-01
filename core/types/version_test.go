package types

import (
	"encoding/binary"
	"github.com/Qitmeer/qitmeer/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	versionData   = []byte{18, 0, 0, 0}   // uint32 => 18
	version1Data  = []byte{18, 0, 17, 28} // uint32 => 470876178
	originVersion = uint32(18)
	version2      = uint32(470876178)
)

func TestVersion(t *testing.T) {
	version := binary.LittleEndian.Uint32(versionData)
	version1 := binary.LittleEndian.Uint32(version1Data)
	assert.NotEqual(t, version1, version)
	assert.Equal(t, version1Data[:2], versionData[:2])
	newVersion := common.GetQitmeerBlockVersion(version2)
	assert.Equal(t, originVersion, newVersion)
}
