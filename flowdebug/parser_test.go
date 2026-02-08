package flowdebug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStringStisfyParser(t *testing.T) {
	testcaseNoerror := []string{
		"result|rest",
		"result.rest",
		"result;rest",
		"result@rest",
	}

	testcastError := []string{
		"@rest",
		"=rest",
		"|rest",
		";rest",
		".rest",
	}

	parse_id := parseStringSatisfy(func(a byte) bool {
		return a != ';' && a != '=' && a != '@' && a != '|' && a != '.'
	}, "dont include {; = @ | .}")

	for _, test := range testcaseNoerror {
		_, _, err := parse_id(test)
		assert.NoError(t, err, test)
	}

	for _, test := range testcastError {
		_, _, err := parse_id(test)
		assert.Error(t, err, test)
	}
}

func TestParseStatsMessageParser(t *testing.T) {
	testcaseNoerror := []string{
		"2tsdf43242.as24sad;a=24245|g;b=342|c;c=235423|c@0.13;d=23423|t@0.0414;",
		"mgv:0.140347410529072;size=0|g;rate=0|g;",
	}
	testcastError := []string{
		"asdfasdf:asd;a=|c@2352.23;",
		"asdfasdf:da.asdfasdf;a=23523|c",
		"asdfasdf:da.asdfasdf;a=adfsadf|c;",
		"dsafsdfsa:dad.asdfas;a=1232|c@b=234|c;",
		"asdfasdf.adsfa;a=124214|c@1351asdf;",
		"a=24245|g;b=342|c;c=235423|c@0.13;d=23423|t@0.0414;",
		"23tsafa.asdfasdf;",
	}

	for _, test := range testcaseNoerror {
		ret1, err := ParseStatsMessage()(test)
		assert.NoError(t, err, test)
		ret2, err := RegexStatsMessage(test)
		assert.NoError(t, err, test)
		assert.Equal(t, ret1, ret2)
	}

	for _, test := range testcastError {
		_, err := ParseStatsMessage()(test)
		assert.Error(t, err, test)
		_, err = RegexStatsMessage(test)
		assert.Error(t, err, test)
	}
}
