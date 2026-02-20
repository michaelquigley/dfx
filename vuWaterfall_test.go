package dfx

import "testing"

func TestSetHistorySize_ClampsInvalidToOne(t *testing.T) {
	w := NewVUWaterfall(2)

	w.SetHistorySize(0)
	if w.HistorySize != 1 {
		t.Fatalf("expected history size '1' after zero input, got '%d'", w.HistorySize)
	}
	if len(w.history) != 1 {
		t.Fatalf("expected internal history length '1', got '%d'", len(w.history))
	}

	w.SetHistorySize(-5)
	if w.HistorySize != 1 {
		t.Fatalf("expected history size '1' after negative input, got '%d'", w.HistorySize)
	}
}

func TestSetLevels_AfterClampMaintainsValidCircularBuffer(t *testing.T) {
	w := NewVUWaterfall(2)
	w.SetHistorySize(0)

	w.SetLevels([]float32{0.5, 0.7})
	if w.historyLen != 1 {
		t.Fatalf("expected historyLen '1', got '%d'", w.historyLen)
	}
	if w.historyHead != 0 {
		t.Fatalf("expected historyHead to wrap to '0', got '%d'", w.historyHead)
	}

	w.SetLevel(1, 0.25)
	if w.historyLen != 1 {
		t.Fatalf("expected historyLen to remain '1', got '%d'", w.historyLen)
	}
	if w.historyHead != 0 {
		t.Fatalf("expected historyHead to remain wrapped at '0', got '%d'", w.historyHead)
	}
}
