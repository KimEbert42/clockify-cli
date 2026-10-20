package clockify

import (
	"encoding/json"
	"testing"
)

func TestFlexibleDurationUnmarshalJSON(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{`3600`, 3600},
		{`"3600"`, 3600},
		{`"PT0S"`, 0},
		{`"PT400H"`, 400 * 3600},
		{`"PT1H30M"`, 5400},
		{`"P1DT2H"`, 26 * 3600},
		{`null`, 0},
		{`""`, 0},
	}
	for _, tc := range cases {
		var d FlexibleDuration
		if err := json.Unmarshal([]byte(tc.in), &d); err != nil {
			t.Fatalf("Unmarshal(%s): %v", tc.in, err)
		}
		if d.Int64() != tc.want {
			t.Errorf("Unmarshal(%s) = %d, want %d", tc.in, d.Int64(), tc.want)
		}
	}
}

func TestProjectUnmarshalRealAPIPayload(t *testing.T) {
	body := `[{"id":"p1","name":"Sample","duration":"PT279H","estimate":{"estimate":"PT400H","type":"MANUAL"},"billable":true}]`
	var projects []Project
	if err := json.Unmarshal([]byte(body), &projects); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len = %d", len(projects))
	}
	p := projects[0]
	if p.Duration.Int64() != 279*3600 {
		t.Errorf("duration = %d", p.Duration.Int64())
	}
	if p.Estimate == nil || p.Estimate.Estimate.Int64() != 400*3600 {
		t.Errorf("estimate = %+v", p.Estimate)
	}
	if p.Estimate.Type != "MANUAL" {
		t.Errorf("estimate type = %q", p.Estimate.Type)
	}
}

func TestEstimateMarshalSeconds(t *testing.T) {
	b, err := json.Marshal(Estimate{Type: "AUTO", Estimate: FlexibleDuration(3600)})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"type":"AUTO","estimate":3600}` {
		t.Errorf("marshal = %s", b)
	}
}
