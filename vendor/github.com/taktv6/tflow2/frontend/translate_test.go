package frontend

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taktv6/tflow2/database"
)

func TestTranslateCondition(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		Key              string
		Value            string
		ExpectedField    int
		ExpectedOperator int
	}{
		{
			Key:              "Timestamp.gt",
			Value:            "1503432000",
			ExpectedField:    database.FieldTimestamp,
			ExpectedOperator: database.OpGreater,
		},
		{
			Key:              "Timestamp.lt",
			Value:            "1503436000",
			ExpectedField:    database.FieldTimestamp,
			ExpectedOperator: database.OpSmaller,
		},
		{
			Key:              "Protocol.eq",
			Value:            "6",
			ExpectedField:    database.FieldProtocol,
			ExpectedOperator: database.OpEqual,
		},
		{
			Key:              "SrcAddr",
			Value:            "1.2.3.4",
			ExpectedField:    database.FieldSrcAddr,
			ExpectedOperator: database.OpEqual,
		},
		{
			Key:              "SrcAs",
			Value:            "5123",
			ExpectedField:    database.FieldSrcAs,
			ExpectedOperator: database.OpEqual,
		},
		{
			Key:              "SrcPfx",
			Value:            "10.8.0.0/16",
			ExpectedField:    database.FieldSrcPfx,
			ExpectedOperator: database.OpEqual,
		},
	}

	fe := Frontend{}
	for _, test := range tests {
		cond, err := fe.translateCondition(test.Key, test.Value)
		assert.NoError(err)
		assert.NotNil(cond)
		assert.Equal(test.ExpectedField, cond.Field)
		assert.Equal(test.ExpectedOperator, cond.Operator)
	}

}

func TestTranslateQuery(t *testing.T) {
	assert := assert.New(t)
	fe := Frontend{}

	query, errors := fe.translateQuery(url.Values{"TopN": []string{"15"}})
	assert.Nil(errors)
	assert.Equal(query.TopN, 15)

	query, errors = fe.translateQuery(url.Values{"Timestamp.lt": []string{"42"}, "Timestamp.gt": []string{"23"}})
	assert.Nil(errors)
	assert.Len(query.Cond, 2)

	query, errors = fe.translateQuery(url.Values{"Unknown": []string{"foo"}})
	assert.EqualError(errors[0], "unknown field: Unknown")
}
