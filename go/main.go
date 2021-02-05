package main

import (
	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"log"
	"syscall/js"
)

func main() {
	log.Println("Starting hcl_wasm")

	// To call go functions from JS, they need to be assigned to global variables. A real application would
	// probably create a single (versioned?) global proxy object, rather than polluting the global namespaces
	js.Global().Set("parse_hcl", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ret := make(map[string]interface{})

		// JS is loosey-goosey, so it's important to check arguments carefully. Panics in go are very annoying, because
		// they cause the go app to exit and JS can no longer call into GO
		if args == nil || len(args) < 3 {
			ret["error"] = js.ValueOf("Not enough args")
			log.Println("Not enough args")
			return ret
		}
		hcl := args[0].String()

		// STDOUT is very helpfully redirected to the console, which makes printf-debugging easy
		log.Println("Parsing HCL", hcl)

		rawContext := args[1]

		context := hcl2.EvalContext{Variables: map[string]cty.Value{}, Functions: map[string]function.Function{}}

		// This exposes a length function to HCL that returns the length of the given string
		context.Functions["length"] = function.New(&function.Spec{
			Params: []function.Parameter{{Type: cty.List(cty.String)}},
			VarParam: nil,
			Type:     function.StaticReturnType(cty.Number),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				length := args[0].LengthInt()
				return cty.NumberIntVal(int64(length)), nil
			},
		})

		// Go/JS communication is very rudimentary. This is the only way to get objects with arbitrary shapes.
		contextKeys := js.Global().Get("Object").Call("keys", rawContext)
		for i := 0; i < contextKeys.Length(); i++ {
			key := contextKeys.Index(i).String()
			rawValue := rawContext.Get(key)

			log.Println("Loaded context value", key, rawValue)

			switch rawValue.Type() {
			case js.TypeString:
				context.Variables[key] = cty.StringVal(rawValue.String())
			case js.TypeNumber:
				context.Variables[key] = cty.NumberFloatVal(rawValue.Float())
			case js.TypeObject:
				vals := make([]cty.Value, 0)

				for i := 0; i < rawValue.Length(); i++ {
					vals = append(vals, cty.StringVal(rawValue.Index(i).String()))
				}

				context.Variables[key] = cty.ListVal(vals)
			default:
				log.Println("Unsupported raw value type", rawValue.Type())
			}

		}

		// Since the go side of things can call any JS functions it can reach, it's possible to manipulate the DOM
		// from go code
		container := args[2]
		for container.Get("firstChild").Truthy() {
			firstChild := container.Get("firstChild")
			log.Println("Removing old element", firstChild)
			container.Call("removeChild", firstChild)
		}

		parser := hclparse.NewParser()
		parsedHCL, diagnostics := parser.ParseHCL([]byte(hcl), "inline.hcl")

		createElements(js.Global().Get("document"), container, parsedHCL.Body, &context)

		// Panics are bad for the reason given above, and shipping non-primitives to JS is painful. Because of that,
		// the best mechanism for surfacing errors to JS is strings + status flags. A real application would have
		// a JS facade over the go code that handles surfacing errors in a pleasant way.
		if diagnostics.HasErrors() {
			ret["error"] = js.ValueOf(diagnostics.Error())
		}

		return ret
	}))

	// Listening to a channel is a convenient way to prevent the application from exiting. Without something like this,
	// the go application exits immediately after being instantiated and there's no way to call the exposed methods
	done := make(chan struct{})
	<-done
}

// This function (and its mutually recursive friend below) handle turning the given HCL structure into DOM elements
func createElements(document js.Value, container js.Value, body hcl2.Body, context *hcl2.EvalContext) {
	supportedTags := []string{"div", "h1", "h2", "h3", "b", "center", "p", "marquee", "span", "content", "ul", "ol", "li", "br"}

	blocks := make([]hcl2.BlockHeaderSchema, 0)
	for _, supportedTag := range supportedTags {
		blocks = append(blocks, hcl2.BlockHeaderSchema{Type: supportedTag})
	}

	blocks = append(blocks, hcl2.BlockHeaderSchema{
		Type:       "dynamic",
		LabelNames: []string{"blockName"},
	})

	schema := hcl2.BodySchema{
		Blocks: blocks,
		Attributes: []hcl2.AttributeSchema{
			{Name: "for_each", Required: false},
		},
	}
	content, _, _ := body.PartialContent(&schema)

	for _, block := range content.Blocks {
		log.Println("Encountered block", block.Type)

		if block.Type == "dynamic" {
			tagName := block.Labels[0]
			log.Println("Dynamic block tag name", tagName)

			attributes, _ := block.Body.JustAttributes()
			value, _ := attributes["for_each"].Expr.Value(context)
			value.ForEachElement(func(key cty.Value, val cty.Value) (stop bool) {
				evalContext := context.NewChild()
				evalContext.Variables = map[string]cty.Value{}
				evalContext.Variables["for_each"] = val
				log.Println("Expanding dynamic block with", tagName, val.GoString())
				handleBlock(tagName, evalContext, block.Body, document, container)
				log.Println("Finished dynamic block expansion")
				return false
			})
		} else {
			handleBlock(block.Type, context, block.Body, document, container)

		}

	}

}

func handleBlock(blockType string, ctx *hcl2.EvalContext, body hcl2.Body, document js.Value, container js.Value) {
	element := document.Call("createElement", blockType)

	attributes, _ := body.JustAttributes()
	for _, attribute := range attributes {
		if attribute.Name != "for_each" {
			value, _ := attribute.Expr.Value(ctx)
			switch value.Type() {
			case cty.String:
				element.Set(attribute.Name, value.AsString())
			case cty.Number:
				f, _ := value.AsBigFloat().Float64()
				element.Set(attribute.Name, f)
			}

			log.Println("Encountered attribute", attribute.Name, value.Type(), value.GoString())
		}
	}

	container.Call("appendChild", element)

	createElements(document, element, body, ctx)
}
