package flowdebug

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type StatsTagMessage struct {
	Tag     string
	HasSign bool
	Value   int
	Ratio   float64
	Metric  string
}

type StatsMessage struct {
	GraphId string
	StatId  string
	TagMsgs []StatsTagMessage
}

func RegexStatsMessage(str string) (*StatsMessage, error) {
	pattern := regexp.MustCompile(`^(?P<id>[^\.;=@\|]*)(\.(?P<stat>[^\.;=@\|]*))?;(?P<tags>.+)$`)

	repeatPattern := regexp.MustCompile(`(?P<tag>[^\.;=@\|]*)(=(?P<value>\d+))?(\|(?P<metric>[^\.;=@\|]+)(@(?P<ratio>[\d.]*))?)?;`)

	match := pattern.FindStringSubmatch(str)
	groupNames := pattern.SubexpNames()

	if len(match) == 0 {
		return nil, fmt.Errorf("invalid pattern: %s", str)
	}

	var stats StatsMessage
	for i, group := range match {
		switch groupNames[i] {
		case "id":
			stats.GraphId = group
		case "stat":
			stats.StatId = group
		case "tags":
			tagGroups := repeatPattern.FindAllStringSubmatch(group, -1)
			tagGroupNames := repeatPattern.SubexpNames()
			var matchedTag strings.Builder
			for _, tagGroup := range tagGroups {
				matchedTag.WriteString(tagGroup[0])
				var tagMsg StatsTagMessage
				tagMsg.Ratio = 1.0
				for j, item := range tagGroup {
					switch tagGroupNames[j] {
					case "tag":
						tagMsg.Tag = item
					case "value":
						if item != "" {
							value, err := strconv.ParseInt(item, 10, 0)
							if err != nil {
								return nil, err
							}
							tagMsg.Value = int(value)
						}
					case "metric":
						tagMsg.Metric = item
					case "ratio":
						if item != "" {
							ratio, err := strconv.ParseFloat(item, 64)
							if err != nil {
								return nil, err
							}
							tagMsg.Ratio = ratio
						}
					}
				}
				stats.TagMsgs = append(stats.TagMsgs, tagMsg)
			}
			if matchedTag.String() != group {
				return nil, fmt.Errorf("invalid tags: %s", group)
			}
		}
	}
	return &stats, nil
}

// id.stat;[tag=value|metric@ratio;].
func (msg *StatsMessage) ToString() string {
	var sb strings.Builder
	sb.WriteString(msg.GraphId)
	sb.WriteString(".")
	sb.WriteString(msg.StatId)
	sb.WriteString(";")
	for i := range msg.TagMsgs {
		tagMsg := msg.TagMsgs[i]
		sb.WriteString(tagMsg.Tag)
		sb.WriteString("=")
		sb.WriteString(strconv.Itoa(tagMsg.Value))
		sb.WriteString("|")
		sb.WriteString(tagMsg.Metric)
		sb.WriteString("@")
		sb.WriteString(strconv.FormatFloat(tagMsg.Ratio, 'f', -1, 32))
		sb.WriteString(";")
	}
	return sb.String()
}

func parseCharIf(tell func(byte) bool, desc string) func(string) (byte, string, error) {
	return func(input string) (byte, string, error) {
		if tell(input[0]) {
			return input[0], input[1:], nil
		} else {
			return 0, input, fmt.Errorf("`%c` is invalid, expected satisfy %s, input: %s", input[0], desc, input)
		}
	}
}

func parseIsChar(c byte) func(string) (byte, string, error) {
	return parseCharIf(func(a byte) bool { return c == a }, fmt.Sprintf("%c", c))
}

/* func parseIsNumber() func(string) (error, byte, string) {
	return parseCharIf(func(a byte) bool { return a >= 48 && a <= 57 }, "in range [0-9]")
} */

func parseStringSatisfy(tell func(byte) bool, desc string) func(string) (string, string, error) {
	charIf := parseCharIf(tell, desc)
	return func(input string) (string, string, error) {
		for i := 0; i < len(input); i++ {
			_, _, err := charIf(input[i:])
			if err != nil {
				if i == 0 {
					return "", input, err
				}
				return input[:i], input[i:], nil
			}
		}
		return input, "", nil
	}
}

func parseString_util(stop byte) func(string) (string, string, error) {
	charIf := parseIsChar(stop)
	return func(input string) (string, string, error) {
		for i := 0; i < len(input); i++ {
			_, rest, err := charIf(input[i:])
			if err == nil {
				return input[:i], rest, nil
			}
		}
		return "", input, fmt.Errorf("there hasn't stop:`%c`", stop)
	}
}

/* func parseNumber() func(string) (error, int, string) {
	stringSatisfy := parseStringSatisfy(func(a byte) bool { return a >= 48 && a <= 57 }, "in range [0-9]")
	return func(input string) (error, int, string) {
		err, ret, rest := stringSatisfy(input)
		if err != nil {
			return err, 0, rest
		}
		iret, err := strconv.Atoi(ret)
		return err, iret, rest
	}
} */

func ParseStatsMessage() func(string) (*StatsMessage, error) {
	parseId := parseStringSatisfy(func(a byte) bool {
		return a != ';' && a != '=' && a != '@' && a != '|' && a != '.'
	}, "dont include {; = @ | .}")
	utilSemicolon := parseString_util(';')
	utilAt := parseString_util('@')
	utilVert := parseString_util('|')
	// id.stat;[tag=value|metric@ratio;]
	return func(input string) (*StatsMessage, error) {
		ret := &StatsMessage{}
		id, rest, err := parseId(input)
		if err != nil {
			return ret, err
		}
		ret.GraphId = id
		_, rest, err = parseIsChar('.')(rest)
		if err != nil {
			return ret, err
		}
		stat, rest, err := parseId(rest)
		if err != nil {
			return ret, err
		}
		ret.StatId = stat
		_, rest, err = parseIsChar(';')(rest)
		if err != nil {
			return ret, err
		}
		if len(rest) == 0 {
			return ret, fmt.Errorf("there hasn't message, input: %s", input)
		}
		ret.TagMsgs = make([]StatsTagMessage, 0, 8)
		for {
			var tagRet StatsTagMessage
			tagRet.Tag, rest, err = parseId(rest)
			if err != nil {
				return ret, err
			}
			_, rest, err = parseIsChar('=')(rest)
			if err != nil {
				return ret, err
			}
			var signAndNumber string
			signAndNumber, rest, err = utilVert(rest)
			if err != nil {
				return ret, err
			}
			if signAndNumber[0] == '+' || signAndNumber[0] == '-' {
				tagRet.HasSign = true
				if signAndNumber[0] == '+' {
					signAndNumber = signAndNumber[1:]
				}
			}
			tagRet.Value, err = strconv.Atoi(signAndNumber)
			if err != nil {
				return ret, err
			}
			var metricAndRatio string
			metricAndRatio, rest, err = utilSemicolon(rest)
			if err != nil {
				return ret, err
			}
			metric, ratio, err := utilAt(metricAndRatio)
			tagRet.Ratio = 1.0
			if err == nil {
				tagRet.Ratio, err = strconv.ParseFloat(ratio, 64)
				if err != nil {
					return ret, err
				}
				tagRet.Metric = metric
			} else {
				tagRet.Metric = metricAndRatio
			}
			ret.TagMsgs = append(ret.TagMsgs, tagRet)
			if len(rest) == 0 {
				break
			}
		}
		return ret, nil
	}
}
