package frontend

import (
	"fmt"
	"net/http"

	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/database"
)

func (fe *Frontend) prometheusHandler(w http.ResponseWriter, r *http.Request) {

	query, errors := fe.translateQuery(r.URL.Query())
	if errors != nil {
		http.Error(w, "Unable to parse query:", 422)
		for _, err := range errors {
			fmt.Fprintln(w, err.Error())
		}
		return
	}

	if query.Breakdown.Count() == 0 {
		http.Error(w, "Breakdown parameter missing. Please pass a comma separated list of:", 422)
		for _, label := range database.GetBreakdownLabels() {
			fmt.Fprintf(w, "- %s\n", label)
		}
		return
	}

	if !query.Cond.Includes(database.FieldTimestamp, database.OpEqual) {
		// Select most recent complete timeslot
		ts := fe.flowDB.CurrentTimeslot() - fe.flowDB.AggregationPeriod()
		query.Cond = append(query.Cond, database.Condition{
			Field:    database.FieldTimestamp,
			Operator: database.OpEqual,
			Operand:  convert.Int64Byte(ts),
		})
	}

	// Run the query
	result, err := fe.flowDB.RunQuery(&query)
	if err != nil {
		http.Error(w, "Query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Empty result?
	if len(result.Timestamps) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Hints for Prometheus
	fmt.Fprintln(w, "# HELP tflow_bytes Bytes transmitted")
	fmt.Fprintln(w, "# TYPE tflow_bytes gauge")

	ts := result.Timestamps[0]

	// Print the data
	for key, val := range result.Data[ts] {
		fmt.Fprintf(w, "tflow_bytes{%s} %d\n", formatBreakdownKey(&key), val)
	}
}

// formats a breakdown key for prometheus
// see tests for examples
func formatBreakdownKey(key *database.BreakdownKey) string {
	return key.Join(`%s="%s"`)
}
