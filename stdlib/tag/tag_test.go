package tag_test

import (
	"testing"

	"github.com/amp-3d/amp-sdk-go/stdlib/tag"
)

func TestTag(t *testing.T) {
	spec := tag.FormSpec(tag.FormSpec(tag.Spec{}, "amp.app"), "meet.galene")
	if spec.Canonic != "amp.app.meet.galene" {
		t.Errorf("FormSpec failed")
	}

	prefix, suffix := spec.LeafTags(2)
	if prefix != "amp.app" || suffix != "meet.galene" {
		t.Errorf("LeafTags failed")
	}
}