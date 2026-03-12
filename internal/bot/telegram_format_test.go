package bot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertMarkdownToTelegramHTML_Headings(t *testing.T) {
	md := "# Heading 1\n## Heading 2"
	expected := "<b>Heading 1</b>\n\n<b>Heading 2</b>"

	result := ConvertMarkdownToTelegramHTML(md)
	assert.Equal(t, expected, result)
}

func TestConvertMarkdownToTelegramHTML_Lists(t *testing.T) {
	md := "- Item 1\n- Item 2\n  - Nested 1\n  - Nested 2\n- Item 3"
	expected := "• Item 1\n• Item 2\n  • Nested 1\n  • Nested 2\n• Item 3"

	result := ConvertMarkdownToTelegramHTML(md)
	assert.Equal(t, expected, result)

	mdOrdered := "1. First\n2. Second"
	expectedOrdered := "1. First\n2. Second"

	resultOrdered := ConvertMarkdownToTelegramHTML(mdOrdered)
	assert.Equal(t, expectedOrdered, resultOrdered)
}

func TestConvertMarkdownToTelegramHTML_CodeBlocks(t *testing.T) {
	md := "Here is some code:\n```go\nfunc main() {}\n```\nAnd `inline` code."

	result := ConvertMarkdownToTelegramHTML(md)
	// It's possible whitespace/newlines might be slightly different depending on goldmark parser,
	// so let's just check the structure.
	assert.True(t, strings.Contains(result, "<pre><code class=\"language-go\">func main() {}"))
	assert.True(t, strings.Contains(result, "<code>inline</code>"))
}

func TestConvertMarkdownToTelegramHTML_MixedFormatting(t *testing.T) {
	md := "This is **bold**, *italic*, and ~~strikethrough~~."
	expected := "This is <b>bold</b>, <i>italic</i>, and <s>strikethrough</s>."
	
	result := ConvertMarkdownToTelegramHTML(md)
	assert.Equal(t, expected, result)
}

func TestConvertMarkdownToTelegramHTML_Links(t *testing.T) {
	md := "Click [here](https://example.com) for more info."
	expected := "Click <a href=\"https://example.com\">here</a> for more info."
	
	result := ConvertMarkdownToTelegramHTML(md)
	assert.Equal(t, expected, result)
}
