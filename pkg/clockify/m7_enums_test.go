package clockify

import "testing"

func TestMapCustomFieldType(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"TEXT", "TXT", false},
		{"text", "TXT", false},
		{"NUMBER", "NUMBER", false},
		{"DROPDOWN", "DROPDOWN_SINGLE", false},
		{"DROPDOWN_SINGLE", "DROPDOWN_SINGLE", false},
		{"DROPDOWN_MULTIPLE", "DROPDOWN_MULTIPLE", false},
		{"CHECKBOX", "CHECKBOX", false},
		{"LINK", "LINK", false},
		{"DATE", "", true},
		{"", "", true},
		{"NOPE", "", true},
	}
	for _, tt := range tests {
		got, err := MapCustomFieldType(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("MapCustomFieldType(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("MapCustomFieldType(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMapInvoiceStatus(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"DRAFT", "UNSENT", false},
		{"FINALIZED", "UNSENT", false},
		{"SENT", "SENT", false},
		{"ACCEPTED", "PAID", false},
		{"PAID", "PAID", false},
		{"VOID", "VOID", false},
		{"UNSENT", "UNSENT", false},
		{"PARTIALLY_PAID", "PARTIALLY_PAID", false},
		{"OVERDUE", "OVERDUE", false},
		{"BOGUS", "", true},
	}
	for _, tt := range tests {
		got, err := MapInvoiceStatus(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("MapInvoiceStatus(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("MapInvoiceStatus(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMapWebhookEvents(t *testing.T) {
	got, err := MapWebhookEvents("TIME_ENTRY_CREATED, NEW_TIME_ENTRY, NEW_PROJECT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %v, want deduped [NEW_TIME_ENTRY NEW_PROJECT]", got)
	}
	if got[0] != "NEW_TIME_ENTRY" || got[1] != "NEW_PROJECT" {
		t.Errorf("got %v", got)
	}

	if _, err := MapWebhookEvents("NOT_AN_EVENT"); err == nil {
		t.Error("expected error for invalid event")
	}
	if _, err := MapWebhookEvents(""); err == nil {
		t.Error("expected error for empty events")
	}
}

func TestMapApprovalState(t *testing.T) {
	got, err := MapApprovalState("approved")
	if err != nil || got != "APPROVED" {
		t.Errorf("MapApprovalState(approved) = %q, %v", got, err)
	}
	if _, err := MapApprovalState("MAYBE"); err == nil {
		t.Error("expected error for invalid state")
	}
}

func TestMapTimeOffStatus(t *testing.T) {
	if got, err := MapTimeOffStatus("PENDING"); err != nil || got != "PENDING" {
		t.Errorf("got %q, %v", got, err)
	}
	if _, err := MapTimeOffStatus("MAYBE"); err == nil {
		t.Error("expected error")
	}
}

func TestValidateTimeOffType(t *testing.T) {
	if err := ValidateTimeOffType("vacation"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidateTimeOffType("PARTY"); err == nil {
		t.Error("expected error")
	}
}

func TestMapApprovalPeriodAndType(t *testing.T) {
	if got, err := MapApprovalPeriod("weekly"); err != nil || got != "WEEKLY" {
		t.Errorf("period = %q, %v", got, err)
	}
	if _, err := MapApprovalPeriod("DAILY"); err == nil {
		t.Error("expected error for DAILY")
	}
	if got, err := MapApprovalType("timesheet"); err != nil || got != "TIMESHEET" {
		t.Errorf("type = %q, %v", got, err)
	}
	if _, err := MapApprovalType("WAT"); err == nil {
		t.Error("expected error")
	}
}

func TestMapSeriesUpdateOption(t *testing.T) {
	if got, err := MapSeriesUpdateOption(""); err != nil || got != "THIS_ONE" {
		t.Errorf("default series = %q, %v", got, err)
	}
	if got, err := MapSeriesUpdateOption("ALL"); err != nil || got != "ALL" {
		t.Errorf("series ALL = %q, %v", got, err)
	}
	if _, err := MapSeriesUpdateOption("MAYBE_ALL"); err == nil {
		t.Error("expected error")
	}
}

func TestMapCustomFieldStatus(t *testing.T) {
	if got, err := MapCustomFieldStatus("visible"); err != nil || got != "VISIBLE" {
		t.Errorf("got %q, %v", got, err)
	}
	if got, err := MapCustomFieldStatus(""); err != nil || got != "" {
		t.Errorf("empty should be ok, got %q, %v", got, err)
	}
	if _, err := MapCustomFieldStatus("HIDDEN"); err == nil {
		t.Error("expected error")
	}
}
