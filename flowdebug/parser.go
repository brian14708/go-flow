package flowdebug

import (
	"fmt"
	"regexp"
	"strconv"
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

func RegexStatsMessage(str string) (error, *StatsMessage) {
	pattern := regexp.MustCompile(`^(?P<id>[^\.;=@\|]*)(\.(?P<stat>[^\.;=@\|]*))?;(?P<tags>.+)$`)

	repeatPattern := regexp.MustCompile(`(?P<tag>[^\.;=@\|]*)(=(?P<value>\d+))?(\|(?P<metric>[^\.;=@\|]+)(@(?P<ratio>[\d.]*))?)?;`)

	match := pattern.FindStringSubmatch(str)
	groupNames := pattern.SubexpNames()

	if len(match) == 0 {
		return fmt.Errorf("Invalid pattern: %s", str), nil
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
			var matchedTag string
			for _, tagGroup := range tagGroups {
				matchedTag += tagGroup[0]
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
								return err, nil
							}
							tagMsg.Value = int(value)
						}
					case "metric":
						tagMsg.Metric = item
					case "ratio":
						if item != "" {
							ratio, err := strconv.ParseFloat(item, 64)
							if err != nil {
								return err, nil
							}
							tagMsg.Ratio = ratio
						}
					}
				}
				stats.TagMsgs = append(stats.TagMsgs, tagMsg)
			}
			if matchedTag != group {
				return fmt.Errorf("Invalid tags: %s", group), nil
			}
		}
	}
	return nil, &stats
}

// id.stat;[tag=value|metric@ratio;]
func (msg *StatsMessage) ToString() string {
	head := msg.GraphId + "." + msg.StatId + ";"
	for i := range msg.TagMsgs {
		tagMsg := msg.TagMsgs[i]
		head += tagMsg.Tag + "=" + strconv.Itoa(tagMsg.Value) + "|" + tagMsg.Metric + "@" + strconv.FormatFloat(tagMsg.Ratio, 'f', -1, 32) + ";"
	}
	return head
}

func parseCharIf(tell func(byte) bool, desc string) func(string) (error, byte, string) {
	return func(input string) (error, byte, string) {
		if tell(input[0]) {
			return nil, input[0], input[1:]
		} else {
			return fmt.Errorf("`%c` is invalid, expected satisfy %s, input: %s", input[0], desc, input), 0, input
		}
	}
}

func parseIsChar(c byte) func(string) (error, byte, string) {
	return parseCharIf(func(a byte) bool { return c == a }, fmt.Sprintf("%c", c))
}

/* func parseIsNumber() func(string) (error, byte, string) {
	return parseCharIf(func(a byte) bool { return a >= 48 && a <= 57 }, "in range [0-9]")
} */

func parseStringSatisfy(tell func(byte) bool, desc string) func(string) (error, string, string) {
	charIf := parseCharIf(tell, desc)
	return func(input string) (error, string, string) {
		for i := 0; i < len(input); i++ {
			err, _, _ := charIf(input[i:])
			if err != nil {
				if i == 0 {
					return err, "", input
				}
				return nil, input[:i], input[i:]
			}
		}
		return nil, input, ""
	}
}

func parseString_util(stop byte) func(string) (error, string, string) {
	charIf := parseIsChar(stop)
	return func(input string) (error, string, string) {
		for i := 0; i < len(input); i++ {
			err, _, rest := charIf(input[i:])
			if err == nil {
				return nil, input[:i], rest
			}
		}
		return fmt.Errorf("There hasnt stop:`%c`", stop), "", input
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

func ParseStatsMessage() func(string) (error, *StatsMessage) {
	parseId := parseStringSatisfy(func(a byte) bool {
		return a != ';' && a != '=' && a != '@' && a != '|' && a != '.'
	}, "dont include {; = @ | .}")
	utilSemicolon := parseString_util(';')
	utilAt := parseString_util('@')
	utilVert := parseString_util('|')
	// id.stat;[tag=value|metric@ratio;]
	return func(input string) (error, *StatsMessage) {
		ret := &StatsMessage{}
		err, id, rest := parseId(input)
		if err != nil {
			return err, ret
		}
		ret.GraphId = id
		err, _, rest = parseIsChar('.')(rest)
		if err != nil {
			return err, ret
		}
		err, stat, rest := parseId(rest)
		if err != nil {
			return err, ret
		}
		ret.StatId = stat
		err, _, rest = parseIsChar(';')(rest)
		if err != nil {
			return err, ret
		}
		if len(rest) == 0 {
			return fmt.Errorf("There hasnt message, input: %s", input), ret
		}
		ret.TagMsgs = make([]StatsTagMessage, 0, 8)
		for {
			var tagRet StatsTagMessage
			err, tagRet.Tag, rest = parseId(rest)
			if err != nil {
				return err, ret
			}
			err, _, rest = parseIsChar('=')(rest)
			if err != nil {
				return err, ret
			}
			var signAndNumber string
			err, signAndNumber, rest = utilVert(rest)
			if err != nil {
				return err, ret
			}
			if signAndNumber[0] == '+' || signAndNumber[0] == '-' {
				tagRet.HasSign = true
				if signAndNumber[0] == '+' {
					signAndNumber = signAndNumber[1:]
				}
			}
			tagRet.Value, err = strconv.Atoi(signAndNumber)
			if err != nil {
				return err, ret
			}
			var metricAndRatio string
			err, metricAndRatio, rest = utilSemicolon(rest)
			if err != nil {
				return err, ret
			}
			err, metric, ratio := utilAt(metricAndRatio)
			tagRet.Ratio = 1.0
			if err == nil {
				tagRet.Ratio, err = strconv.ParseFloat(ratio, 64)
				if err != nil {
					return err, ret
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
		return nil, ret
	}
}
