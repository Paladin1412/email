package RFC2047_test

// =?UTF-8?B?Y2hlLmJhaWR1LmNvbS3nmb7luqbnu5/orqHotovlir/liIbmnpDmiqXlkYpf?=  =?UTF-8?B?MjAxNDA5MDQtMjAxNDA5MDQ=?=

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"../RFC2047"
)

func TestDecode(t *testing.T) {
	input := "=?UTF-8?B?Y2hlLmJhaWR1LmNvbS3nmb7luqbnu5/orqHotovlir/liIbmnpDmiqXlkYpf?=  =?UTF-8?B?MjAxNDA5MDQtMjAxNDA5MDQ=?="
	assert.Equal(t, "che.baidu.com-百度统计趋势分析报告_20140904-20140904",
		RFC2047.Decode(input))
	assert.Equal(t, "Re: "+input,
		RFC2047.Decode("Re: "+input))
}