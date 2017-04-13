package annotations


import (
	"testing"
	"github.com/stretchr/testify/assert"
	"reflect"
	"fmt"
)
const (
	MENTIONS = "MENTIONS"
	MAJOR_MENTIONS = "MAJOR_MENTIONS"
	ABOUT = "ABOUT"
	IS_CLASSIFIED_BY = "IS_CLASSIFIED_BY"
	IS_PRIMARILY_CLASSIFIED_BY = "IS_PRIMARILY_CLASSIFIED_BY"
	HAS_AUTHOR = "HAS_AUTHOR"
	ConceptA = "3a2359b1-9326-4b80-9b97-2a91ccd68d23"
	ConceptB = "df1fead1-5e99-4e92-b23d-fb3cee7f17f2"
)
// Test case definitions taken from https://www.lucidchart.com/documents/edit/df1fead1-5e99-4e92-b23d-fb3cee7f17f2/1?kme=Clicked%20E-mail%20Link&kmi=julia.fernee@ft.com&km_Link=DocInviteButton&km_DocInviteUserArm=T-B
var tests = []struct {
	input            []annotation
	expectedOutput   []annotation
	testDesc string
} {

	{
		[]annotation {
			{ Predicate: MENTIONS, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: MENTIONS, ID: ConceptA, },
		},
		"1. Returns one occurance of Mentions for this concept",
	},
	{
		[]annotation {
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
		},
		"2. Returns one occurance of Major Mentions for this concept",
	},
	{
		[]annotation {
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		"3. Returns one occurance of About for this concept",
	},
	{
		[]annotation {
			{ Predicate: MENTIONS, ID: ConceptA, },
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		"4. Returns one occurance of About for this concept",
	},
	{
		[]annotation {
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptA, },
		},
		"5. Returns one occurance of Is Classified By for this concept",
	},
	{
		[]annotation {
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
		},
		"6. Returns one occurance of Is Primarily Classified By for this concept",
	},
	{
		[]annotation {

			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
			{ Predicate: HAS_AUTHOR, ID: ConceptA, },

		},
		[]annotation {
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
			{ Predicate: HAS_AUTHOR, ID: ConceptA, },
		},
		"7. Returns one occurance returns Has Author & Major Mentions for this concept",
	},
	{
		[]annotation {

			{ Predicate: ABOUT, ID: ConceptA, },
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
			{ Predicate: HAS_AUTHOR, ID: ConceptA, },

		},
		[]annotation {
			{ Predicate: ABOUT, ID: ConceptA, },
			{ Predicate: HAS_AUTHOR, ID: ConceptA, },
		},
		"8. Returns Has Author & About for this concept",
	},
	{
		[]annotation {

			{ Predicate: ABOUT, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		"9. Returns About for this concept",
	},
	{
		[]annotation {
			{ Predicate: MENTIONS, ID: ConceptA, },
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: ABOUT, ID: ConceptA, },
		},
		"10. Returns About for this concept",
	},
	{
		[]annotation {
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
		},
		[]annotation {
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
		},
		"11. Returns one occurance of Is Primarily Classified By for this concept",
	},
	{
		[]annotation {
			{ Predicate: MAJOR_MENTIONS, ID: ConceptA, },
			{ Predicate: ABOUT, ID: ConceptA, },
			{ Predicate: MENTIONS, ID: ConceptB, },
		},
		[]annotation {
			{ Predicate: ABOUT, ID: ConceptA, },
			{ Predicate: MENTIONS, ID: ConceptB, },
		},
		"12. Returns ABOUT annotation for one concept and MENTIONS annotations for anothr",
	},
	{
		[]annotation {
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptB, },
		},
		[]annotation {
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptB, },
		},
		"13. Returns IS_PRIMARILY_CLASSIFIED_BY annotation for one concept and IS_CLASSIFIED_BY annotations for anothr",
	},
	{
		[]annotation {
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptB, },
		},
		[]annotation {
			{ Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA, },
			{ Predicate: IS_CLASSIFIED_BY, ID: ConceptB, },
		},
		"13. Returns IS_PRIMARILY_CLASSIFIED_BY annotation for one concept and IS_CLASSIFIED_BY annotations for anothr",
	},
}


func TestFilterForBasicSingleConcept(t *testing.T) {
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s",  test.testDesc), func(t *testing.T) {
			filter := NewAnnotationsFilter()
			for _, a := range test.input {
				filter.Add(a)
			}
			actualOutput := filter.Filter()
			assert.True(t, reflect.DeepEqual(test.expectedOutput, actualOutput),
				fmt.Sprintf("Expected %d annotations but returned %d.", len(test.expectedOutput), len(actualOutput)))
		})
	}
}


