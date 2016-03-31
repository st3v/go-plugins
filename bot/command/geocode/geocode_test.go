package geocode

import (
	"testing"
)

func TestGeocode(t *testing.T) {
	testData := []struct {
		address  string
		response string
	}{
		{"somerset house", "51.511028,-0.117194"},
	}

	command := Geocode()

	for _, d := range testData {
		rsp, err := command.Exec("geocode", d.address)
		if err != nil {
			t.Fatal(err)
		}

		if string(rsp) != d.response {
			t.Fatal("Expected %s got %s", d.response, string(rsp))

		}
	}
}
