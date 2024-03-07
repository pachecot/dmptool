package ref

import "testing"

func TestParseRef(t *testing.T) {
	s := `	HwPumpStatus = NUSite\EllHall\Hx1HwPumpStatus`
	r := parseRefs(s)

	if len(r) != 1 {
		t.Error("expected a reference")
	}
	if r[0] != `NUSite\EllHall\Hx1HwPumpStatus` {
		t.Errorf(`expected a 'NUSite\EllHall\Hx1HwPumpStatus' got '%s'`, r[0])
	}
}

func TestParseRef1(t *testing.T) {
	s := `	HwPumpStatus = s(NUSite\EllHall\Hx1HwPumpStatus,NUSite\EllHall\Hx1HwPumpStatus)`
	r := parseRefs(s)

	if len(r) != 2 {
		t.Error("expected a reference")
	}
	if r[0] != `NUSite\EllHall\Hx1HwPumpStatus` {
		t.Errorf(`expected a 'NUSite\EllHall\Hx1HwPumpStatus' got '%s'`, r[0])
	}
	if r[1] != `NUSite\EllHall\Hx1HwPumpStatus` {
		t.Errorf(`expected a 'NUSite\EllHall\Hx1HwPumpStatus' got '%s'`, r[1])
	}
}

func TestParseRef2(t *testing.T) {
	s := `        SpaceSptBaseDegC = round((5 / 9) * (SpaceSptBase - 32)) 'nice simple ##`
	r := parseRefs(s)

	if len(r) != 0 {
		t.Errorf("didn't expect any references. got %s", r[0])
	}
}
