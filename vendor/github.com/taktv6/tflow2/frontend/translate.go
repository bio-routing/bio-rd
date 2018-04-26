package frontend

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/database"
)

func (fe *Frontend) translateCondition(field, value string) (*database.Condition, error) {
	var operatorStr string

	// Extract operator if included in field name
	i := strings.IndexRune(field, '.')
	if i > 0 {
		operatorStr = field[i+1:]
		field = field[:i]
	}

	var operand []byte
	var operator int
	fieldNum := database.GetFieldByName(field)

	switch fieldNum {
	case database.FieldTimestamp:
		op, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		operand = convert.Int64Byte(int64(op))

	case database.FieldProtocol:
		id, err := strconv.Atoi(value)
		operand = convert.Uint8Byte(uint8(id))
		if err != nil {
			protocolsByName := fe.iana.GetIPProtocolsByName()
			operand = convert.Uint8Byte(protocolsByName[value])
		}

	case database.FieldSrcPort, database.FieldDstPort, database.FieldIntIn, database.FieldIntOut:
		op, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		operand = convert.Uint16Byte(uint16(op))

	case database.FieldSrcAddr, database.FieldDstAddr, database.FieldNextHop:
		operand = convert.IPByteSlice(value)

	case database.FieldSrcAs, database.FieldDstAs, database.FieldNextHopAs:
		op, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		operand = convert.Uint32Byte(uint32(op))

	case database.FieldSrcPfx, database.FieldDstPfx:
		_, pfx, err := net.ParseCIDR(string(value))
		if err != nil {
			return nil, err
		}
		operand = []byte(pfx.String())

	case database.FieldIntInName, database.FieldIntOutName, database.FieldAgent:
		operand = []byte(value)

	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}

	switch operatorStr {
	case "eq", "":
		operator = database.OpEqual
	case "ne":
		operator = database.OpUnequal
	case "gt":
		operator = database.OpGreater
	case "lt":
		operator = database.OpSmaller
	default:
		return nil, fmt.Errorf("invalid operator: %s", operatorStr)
	}

	return &database.Condition{
		Field:    fieldNum,
		Operator: operator,
		Operand:  operand,
	}, nil
}

// translateQuery translates URL parameters to the internal represenation of a query
func (fe *Frontend) translateQuery(params url.Values) (q database.Query, errors []error) {
	for key, values := range params {
		var err error
		value := values[0]
		switch key {
		case "TopN":
			q.TopN, err = strconv.Atoi(value)
		case "Breakdown":
			err = q.Breakdown.Set(strings.Split(value, ","))
		default:
			var cond *database.Condition
			cond, err = fe.translateCondition(key, value)
			if cond != nil {
				q.Cond = append(q.Cond, *cond)
			}
		}
		if err != nil {
			errors = append(errors, err)
		}
	}

	return
}
