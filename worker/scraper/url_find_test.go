package scraper

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestCase struct {
	Desc    string
	Input   string
	Results []string
	Fn      func([]byte) []string
}

var testCases []TestCase = []TestCase{
	TestCase{
		Desc: "HTML Doc URLs",
		Fn:   findHTMLDocURLs,
		Input: `
<a href="some-URL">url</a>
<img src="some-url2"/>
<a class="gb_f" href="https://www.google.com/imghp?hl=en&amp;tab=wi&amp;authuser=0" data-pid="2">Images</a>
`,
		Results: []string{
			"some-URL",
			"some-url2",
			"https://www.google.com/imghp?hl=en&amp;tab=wi&amp;authuser=0",
		},
	},

	TestCase{
		Desc: "CSS Doc URLs",
		Fn:   findCSSDocURLs,
		Input: `
background-css: url('something.png');
background-css: url("http://www.example.com/other.jpg");
`,
		Results: []string{
			"something.png",
			"http://www.example.com/other.jpg",
		},
	},

	TestCase{
		Desc: "Gengeric Doc URLs",
		Fn:   findGenericDocURLs,
		Input: `
// some comment
"https://www.example.com/",
var url = '//example.com'
`,
		Results: []string{
			"https://www.example.com/",
			"//example.com",
		},
	},
}

func TestFindURLs(t *testing.T) {
	for i, c := range testCases {
		urls := c.Fn([]byte(c.Input))
		require.Equal(t, len(c.Results), len(urls), "%s:%d: Expect number of results to match", c.Desc, i)
		for j, u := range urls {
			assert.Equal(t, c.Results[j], u, "%s:%d:%d: Expected URL found to match", c.Desc, i, j)
		}
	}
}
