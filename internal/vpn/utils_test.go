package vpn

import "testing"

func TestCountry2Flag(t *testing.T) {
	f, err := country2flag("us")
	if f != "&#x1F1FA;&#x1F1F8;" || err != nil {
		t.Error()
	}

	f, err = country2flag("US")
	if f != "&#x1F1FA;&#x1F1F8;" || err != nil {
		t.Error()
	}

	f, err = country2flag("al")
	if f != "&#x1F1E6;&#x1F1F1;" || err != nil {
		t.Error()
	}

	f, err = country2flag("aL")
	if f != "&#x1F1E6;&#x1F1F1;" || err != nil {
		t.Error()
	}

	f, err = country2flag("az")
	if f != "&#x1F1E6;&#x1F1FF;" || err != nil {
		t.Error()
	}

	f, err = country2flag("Az")
	if f != "&#x1F1E6;&#x1F1FF;" || err != nil {
		t.Error()
	}

	_, err = country2flag("u")
	if err == nil {
		t.Error()
	}

	_, err = country2flag("uuu")
	if err == nil {
		t.Error()
	}

	_, err = country2flag("11")
	if err == nil {
		t.Error()
	}

	_, err = country2flag("u@")
	if err == nil {
		t.Error()
	}

	_, err = country2flag("øu")
	if err == nil {
		t.Error()
	}

	_, err = country2flag("☃u")
	if err == nil {
		t.Error()
	}
}
