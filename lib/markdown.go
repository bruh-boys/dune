package lib

import (
	"math/rand"
	"time"

	"github.com/dunelang/dune"

	"github.com/russross/blackfriday"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	dune.RegisterLib(Markdown, `

declare namespace markdown {

    export function toHTML(n: string | byte[]): string
}

`)
}

var Markdown = []dune.NativeFunction{
	{
		Name:      "markdown.toHTML",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			b := args[0].ToBytes()
			out := blackfriday.MarkdownCommon(b)
			return dune.NewString(string(out)), nil
		},
	},
}
