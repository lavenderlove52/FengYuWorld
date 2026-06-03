package util

import "testing"

func TestSlugifyAllNumeric(t *testing.T) {
	if got := Slugify("123"); got != "123-post" {
		t.Fatalf("Slugify(123) = %q, want 123-post", got)
	}
}

func TestSlugIsAllNumeric(t *testing.T) {
	if !SlugIsAllNumeric("123") {
		t.Fatal("expected 123 to be all numeric")
	}
	if SlugIsAllNumeric("123-abc") {
		t.Fatal("123-abc should not be all numeric")
	}
}
