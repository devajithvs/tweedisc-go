package twitter

import "testing"

func TestGetTweetID(t *testing.T) {
	tables := []struct {
		input  string
		output string
	}{
		{"https://twitter.com/elonmusk/status/1516620752040116231?s=20&t=KY-gboBom_qLR0OdfpHcgQ", "1516620752040116231"},
	}

	for _, table := range tables {
		output := GetTweetID(table.input)
		if output != table.output {
			t.Errorf("Tweet id for (%s) was incorrect, got: %s but want: %s", table.input, output, table.output)
		}
	}

}
