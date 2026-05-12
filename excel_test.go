package document

import (
	"strings"
	"testing"

	"apigo.cc/go/file"
)

func TestExcel_Basic(t *testing.T) {
	filename := "test_basic.xlsx"
	defer file.Remove(filename)

	xls := NewExcel()
	table := [][]any{
		{"Name", "Age", "City"},
		{"Alice", 25, "New York"},
		{"Bob", 30, "London"},
	}

	err := xls.Set("Sheet1", table, "A1", "")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	err = xls.Save(filename)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	doc2, err := Open(filename)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	xls2 := doc2.(*Excel)

	data, err := xls2.Get("Sheet1", "A1", "C3")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(data) != 3 || len(data[1]) != 3 {
		t.Errorf("Unexpected data length: %v", data)
	}

	if castTo[string](data[1][0]) != "Alice" {
		t.Errorf("Expected Alice, got %v", data[1][0])
	}
}

func TestExcel_Data(t *testing.T) {
	filename := "test_data.xlsx"
	defer file.Remove(filename)

	xls := NewExcel()
	data := []map[string]any{
		{"ID": 1, "Value": "A"},
		{"ID": 2, "Value": "B"},
	}

	err := xls.SetData("Data", data, "A1", "")
	if err != nil {
		t.Fatalf("SetData failed: %v", err)
	}

	err = xls.Save(filename)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	doc2, err := Open(filename)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	xls2 := doc2.(*Excel)

	readData, err := xls2.GetData("Data", "A1", "")
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}

	if len(readData) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(readData))
	}

	if castTo[string](readData[0]["Value"]) != "A" {
		t.Errorf("Expected A, got %v", readData[0]["Value"])
	}
}

func TestUnifiedAPI(t *testing.T) {
	// Test Excel with Unified API
	filename := "test_unified.xlsx"
	defer file.Remove(filename)

	doc, err := Open(filename)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if xls, ok := doc.(*Excel); ok {
		xls.SetData("Sheet1", []map[string]any{{"Name": "Unified", "Value": 100}}, "A1", "")
	}

	err = doc.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify JSON/Markdown
	jsonStr := doc.ToJSON()
	if !strings.Contains(jsonStr, "Unified") {
		t.Errorf("ToJSON failed: %s", jsonStr)
	}

	mdStr := doc.ToMarkdown()
	if !strings.Contains(mdStr, "Unified") {
		t.Errorf("ToMarkdown failed: %s", mdStr)
	}
}

func TestIDGeneration(t *testing.T) {
	tests := []struct {
		col, row int
		expected string
	}{
		{0, 0, "A1"},
		{25, 0, "Z1"},
		{26, 0, "AA1"},
		{0, 9, "A10"},
	}

	for _, tt := range tests {
		got := MakeCellID(tt.col, tt.row)
		if got != tt.expected {
			t.Errorf("MakeCellID(%d, %d) = %s; want %s", tt.col, tt.row, got, tt.expected)
		}
		
		c, r := ParseCellID(tt.expected)
		if c != tt.col || r != tt.row {
			t.Errorf("ParseCellID(%s) = (%d, %d); want (%d, %d)", tt.expected, c, r, tt.col, tt.row)
		}
	}
}

// Helper for test assertions
func castTo[T any](v any) T {
	var res T
	if val, ok := v.(T); ok {
		return val
	}
	return res
}
