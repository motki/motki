package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/motki/motkid/cli"
	"github.com/motki/motkid/cli/text"
	"github.com/motki/motkid/evedb"
	"github.com/motki/motkid/log"
	"github.com/motki/motkid/model"
	"github.com/shopspring/decimal"
)

// ProductCommand provides an interactive manager for production chains.
type ProductCommand struct {
	env    *cli.Prompter
	model  *model.Manager
	evedb  *evedb.EveDB
	logger log.Logger
}

func NewProductCommand(p *cli.Prompter, evedb *evedb.EveDB, mdl *model.Manager, logger log.Logger) ProductCommand {
	return ProductCommand{p, mdl, evedb, logger}
}

func (c ProductCommand) Prefixes() []string {
	return []string{"product", "prod"}
}

func (c ProductCommand) Description() string {
	return "Manipulate production chains."
}

func (c ProductCommand) Handle(subcmd string, args ...string) {
	switch {
	case len(subcmd) == 0:
		c.PrintHelp()

	case subcmd == "new" || subcmd == "add" || subcmd == "create":
		c.newProduct(args...)

	case subcmd == "show":
		c.showProduct(args...)

	case subcmd == "list":
		c.listProducts()

	case subcmd == "edit":
		c.editProduct(args...)

	case subcmd == "view" || subcmd == "preview":
		c.previewProduct(args...)

	default:
		fmt.Printf("Unknown subcommand: %s\n", subcmd)
		c.PrintHelp()
	}
}

func (c ProductCommand) PrintHelp() {
	colWidth := 20
	fmt.Println(text.WrapText(`Command "product" can be used to manipulate production chains.

All production chains used here have the corporation ID of 0, and are therefore not accessible by normal means.

When invoking a subcommand, if the optional ID parameter is omitted, an interactive prompt will begin to collect the necessary details.`, text.StandardTerminalWidthInChars))
	fmt.Printf(`
Subcommands:
  %s Preview production chains for a specific item type.
  %s Create a new production chain.

  %s List all production chains for corpID 0.
  %s Display details for a given production chain.
  %s Edit an existing production chain.
`,
		text.PadTextRight("view [typeID]", colWidth),
		text.PadTextRight("add [typeID]", colWidth),
		text.PadTextRight("list", colWidth),
		text.PadTextRight("show [productID]", colWidth),
		text.PadTextRight("edit [productID]", colWidth))
}

// getProductName returns the given product's name.
func (c ProductCommand) getProductName(p *model.Product) string {
	t, err := c.evedb.GetItemType(p.TypeID)
	if err != nil {
		c.logger.Debugf("unable to get item name: %s", err.Error())
		return "[Error]"
	}
	return t.Name
}

// getRegionName returns the given region's name.
func (c ProductCommand) getRegionName(regionID int) string {
	r, err := c.evedb.GetRegion(regionID)
	if err != nil {
		c.logger.Debugf("unable to get region name: %s", err.Error())
		return "[Error]"
	}
	return r.Name
}

// printProductInfo prints production chain details.
func (c ProductCommand) printProductInfo(p *model.Product) {
	batchSize := decimal.NewFromFloat(float64(p.BatchSize))
	costEach := p.Cost().Mul(batchSize) // Cost has quantity baked in.
	batchQuantity := decimal.NewFromFloat(float64(p.Quantity)).Mul(batchSize)
	sellEach := p.MarketPrice.Mul(batchQuantity)
	profitEach := sellEach.Sub(costEach)
	marginEach := decimal.Zero
	if sellEach.Cmp(decimal.Zero) != 0 {
		marginEach = profitEach.Div(sellEach).Mul(decimal.NewFromFloat(100))
	}
	unitLabel := "unit"
	if batchQuantity.GreaterThan(decimal.NewFromFloat(1)) {
		unitLabel = fmt.Sprintf("%s units", batchQuantity)
	}
	fmt.Println(text.CenterText(c.getProductName(p), text.StandardTerminalWidthInChars))
	fmt.Println(text.CenterText(c.getRegionName(p.MarketRegionID), text.StandardTerminalWidthInChars))
	fmt.Println()
	fmt.Printf(
		" #  %s%s%s%s\n",
		text.PadTextRight("Material Name", 29),
		text.PadTextLeft("Cost/ea", 17),
		text.PadTextLeft("Qty Req", 12),
		text.PadTextLeft("Cost/"+unitLabel, 19))
	index := new(int)
	for _, part := range p.Materials {
		c.printChildProductInfo(part, batchSize, p.MaterialEfficiency, index, 0)
	}
	fmt.Println()
	fmt.Printf("%s%s%s\n", text.PadTextLeft(fmt.Sprintf("Per %s", unitLabel), 50), text.PadTextLeft("Revenue", 12), text.PadCurrencyLeft(sellEach, 19))
	fmt.Printf("%s%s%s\n", text.PadTextLeft(fmt.Sprintf("%s%% ME", p.MaterialEfficiency.Mul(decimal.NewFromFloat(100)).StringFixed(0)), 50), text.PadTextLeft("Cost", 12), text.PadCurrencyLeft(costEach, 19))
	fmt.Printf("%s%s\n", text.PadTextLeft("Profit", 61), text.PadCurrencyLeft(profitEach, 19))
	fmt.Printf("%s%s\n", text.PadTextLeft("Margin", 61), "      %"+text.PadTextLeft(marginEach.StringFixed(2), 12))

	fmt.Println()
	fmt.Println("* 'M' indicates the component will be produced in-house.")
	fmt.Println()
}

// printChildProductInfo displays a single component's details.
//
// This function calls itself recursively to traverse the entire production
// chain.
func (c ProductCommand) printChildProductInfo(p *model.Product, parentBatchSize decimal.Decimal, parentME decimal.Decimal, index *int, indent int) {
	*index += 1
	costEach := p.Cost()
	qtyAfterME := decimal.NewFromFloat(float64(p.Quantity)).Mul(parentBatchSize).
		Div(decimal.NewFromFloat(1).Add(parentME)).Round(0)
	costTotal := p.Cost().Mul(qtyAfterME)

	var kind string
	if p.Kind == model.ProductManufacture {
		kind = "M"
	}
	fmt.Printf(
		"%s  %s%s%s%s%s\n",
		text.PadTextLeft(strconv.Itoa(*index), 3),
		text.PadTextRight(strings.Repeat("  ", indent)+c.getProductName(p), 30),
		text.PadTextLeft(kind, 2),
		text.PadCurrencyLeft(costEach, 15),
		text.PadIntegerLeft(int(qtyAfterME.IntPart()), 12),
		text.PadCurrencyLeft(costTotal, 19))
	return
	indent += 1
	if p.Kind == model.ProductManufacture {
		for _, part := range p.Materials {
			c.printChildProductInfo(part, parentBatchSize, p.MaterialEfficiency, index, indent)
		}
	}
}

func (c ProductCommand) getProductLineIndex(p *model.Product) map[int]*model.Product {
	index := map[int]*model.Product{}
	curr := 0
	index[curr] = p // Root is 0.
	for _, part := range p.Materials {
		curr += 1
		index[curr] = part
	}
	return index
}

// newProduct creates a new production chain and opens it in the editor.
func (c ProductCommand) newProduct(args ...string) {
	if p := c.previewProduct(args...); p != nil {
		c.productEditor(p)
	}
}

// editProduct loads a production chain and opens it in the editor.
func (c ProductCommand) editProduct(args ...string) {
	productID := 0
	var ok bool
	var err error
	if len(args) > 0 {
		productID, err = strconv.Atoi(args[0])
	}
	if err != nil || productID <= 0 {
		productID, ok = c.env.PromptInt("Specify Product ID", nil, validateIntGreaterThan(0))
		if !ok {
			return
		}
	}
	product, err := c.model.GetProduct(0, productID)
	if err != nil {
		c.logger.Debugf("unable to load production chain: %s", err.Error())
		fmt.Println("Error loading production chain from db, try again.")
		return
	}
	c.printProductInfo(product)
	c.productEditor(product)
}

const defaultMarketRegionID = 10000043 // Domain, so Amarr.

// previewProduct displays a default view for a given typeID.
func (c ProductCommand) previewProduct(args ...string) *model.Product {
	item, ok := c.env.PromptItemTypeDetail("Specify Item Type", strings.Join(args, " "))
	if !ok {
		return nil
	}
	product, err := c.model.NewProduct(0, item.ID)
	if err != nil {
		c.logger.Warnf("unable to create product: %s", err.Error())
		fmt.Println("Error creating production chain, try again.")
		return nil
	}
	for _, mat := range product.Materials {
		if len(mat.Materials) > 0 {
			mat.Kind = model.ProductManufacture
		}
	}
	if err = c.model.UpdateProductMarketPrices(product, defaultMarketRegionID); err != nil {
		c.logger.Warnf("unable to populate production chain prices: %s", err.Error())
		fmt.Println("Error loading production chain market prices, try again.")
		return nil
	}
	c.printProductInfo(product)
	return product
}

// showProduct loads and displays a production chain's details.
func (c ProductCommand) showProduct(args ...string) {
	productID := 0
	var ok bool
	var err error
	if len(args) > 0 {
		productID, err = strconv.Atoi(args[0])
	}
	if err != nil || productID <= 0 {
		productID, ok = c.env.PromptInt("Specify Product ID", nil, validateIntGreaterThan(0))
		if !ok {
			return
		}
	}
	product, err := c.model.GetProduct(0, productID)
	if err != nil {
		c.logger.Debugf("unable to load product: %s", err.Error())
		fmt.Println("Error loading production chain from db, try again.")
		return
	}
	c.printProductInfo(product)
}

// listProducts lists all the production chains in corpID 0.
func (c ProductCommand) listProducts() {
	products, err := c.model.GetAllProducts(0)
	if err != nil {
		c.logger.Debugf("unable to fetch production chain: %s", err.Error())
		fmt.Println("Error loading production chain from db, try again.")
		return
	}
	fmt.Println("Listing", len(products), "production chains.")
	fmt.Println()
	if len(products) == 0 {
		fmt.Println("There are no production chains. Create a new production chain with")
		fmt.Println("  product add")
		return
	}
	fmt.Printf(
		"%s%s%sType ID\n",
		text.PadTextRight("ID", 12),
		text.PadTextRight("Region", 12),
		text.PadTextRight("Name", 42))
	for _, prod := range products {
		fmt.Printf(
			"%-12.f%s%s%d\n",
			float64(prod.ProductID),
			text.PadTextRight(c.getRegionName(prod.MarketRegionID), 12),
			text.PadTextRight(c.getProductName(prod), 42),
			prod.TypeID)
	}
}

// productEditor starts an interactive session for managing the given production chain.
func (c ProductCommand) productEditor(p *model.Product) {
	lineIndex := c.getProductLineIndex(p)
	shownLineNumberHint := false
	var validLineNumber = func(val int) (int, bool) {
		_, ok := lineIndex[val]
		if !ok {
			fmt.Printf("Invalid line number %d.\n", val)
		}
		return val, ok
	}
	var promptLineNumber = func(prompt string, initVal string, filters ...func(int) (int, bool)) (*model.Product, bool) {
		if v, err := strconv.Atoi(initVal); err == nil {
			if line, ok := validLineNumber(v); ok {
				return lineIndex[line], true
			}
		}
		if !shownLineNumberHint {
			fmt.Println(text.WrapText(fmt.Sprintf("Hint: line 0 is the main item, %s.\n", c.getProductName(p)), text.StandardTerminalWidthInChars))
			shownLineNumberHint = true
		}
		line, ok := c.env.PromptInt(prompt, nil, validLineNumber)
		if !ok {
			return nil, false
		}
		// Presume the line exists since promptInt filtered it already.
		return lineIndex[line], true
	}
	for {
		cmd, args, ok := c.env.PromptStringWithArgs(
			"Specify operation [Q,S,V,D,U,R,C,B,F,M,P,?]",
			nil,
			transformStringToCaps,
			validateStringIsOneOf([]string{"Q", "S", "V", "D", "U", "R", "C", "B", "F", "M", "P", "?"}))
		cmd = strings.ToUpper(cmd)
		if !ok || cmd == "Q" {
			return
		}
		var firstArg string
		if len(args) > 0 {
			firstArg = args[0]
		}
		switch cmd {
		case "S":
			if err := c.model.SaveProduct(p); err != nil {
				c.logger.Warnf("unable to save production chain: %s", err.Error())
				fmt.Println("Error saving production chain, try again.")
				continue
			}
			fmt.Println("Production chain saved.")
			return

		case "D":
			prod, ok := promptLineNumber("Show detail for which line", firstArg)
			if !ok {
				continue
			}
			fmt.Printf("Showing detail for %s.\n\n", c.getProductName(prod))
			c.printProductInfo(prod)
			fmt.Println("Enter Q or S to return to the previous product.")
			c.productEditor(prod)
			fmt.Printf("Returned to detail for %s\n", c.getProductName(p))

		case "C":
			prod, ok := promptLineNumber("Edit cost for which line", firstArg)
			if !ok {
				continue
			}
			prodName := c.getProductName(prod)
			val, ok := c.env.PromptDecimal(fmt.Sprintf("Enter new cost per unit for %s (current: %s)", prodName, prod.Cost()), nil)
			if !ok {
				continue
			}
			prod.MarketPrice = val
			fmt.Printf("Updated %s per unit cost to %s.\n", prodName, prod.Cost())

		case "F":
			prod, ok := promptLineNumber("Edit material efficiency for which line", firstArg)
			if !ok {
				continue
			}
			prodName := c.getProductName(prod)
			val, ok := c.env.PromptDecimal(fmt.Sprintf("Enter new material efficiency for %s (current: %s)", prodName, prod.MaterialEfficiency), nil)
			if !ok {
				continue
			}
			prod.MaterialEfficiency = val
			fmt.Printf("Updated %s material efficiency to %s.\n", prodName, prod.MaterialEfficiency)

		case "M":
			prod, ok := promptLineNumber("Edit production mode for which line", firstArg)
			if !ok {
				continue
			}
			prodName := c.getProductName(prod)
			val, ok := c.env.PromptString(fmt.Sprintf("Enter new mode for %s (current: %s)", prodName, prod.Kind), nil, validateStringIsOneOf([]string{"buy", "build"}))
			if !ok {
				continue
			}
			if val == "buy" {
				prod.Kind = model.ProductBuy
			} else {
				prod.Kind = model.ProductManufacture
			}
			fmt.Printf("Updated %s production mode to %s.\n", prodName, prod.Kind)

		case "B":
			prod, ok := promptLineNumber("Edit cost for which line", firstArg)
			if !ok {
				continue
			}
			prodName := c.getProductName(prod)
			val, ok := c.env.PromptInt(fmt.Sprintf("Enter new batch size for %s (current: %d)", prodName, prod.BatchSize), nil, validateIntGreaterThan(0))
			if !ok {
				continue
			}
			prod.BatchSize = val
			fmt.Printf("Updated %s batch size to %d.\n", prodName, prod.BatchSize)

		case "P":
			prodName := c.getProductName(p)
			val, ok := c.env.PromptDecimal(fmt.Sprintf("Enter new sell price for %s (current: %s)", prodName, p.MarketPrice), nil)
			if !ok {
				continue
			}
			p.MarketPrice = val
			fmt.Printf("Updated %s sell price to %s.\n", prodName, p.MarketPrice)

		case "V":
			fmt.Println()
			c.printProductInfo(p)

		case "U":
			if err := c.model.UpdateProductMarketPrices(p, p.MarketRegionID); err != nil {
				c.logger.Errorf("unable to fetch market prices for region %d: %s", p.MarketRegionID, err.Error())
				fmt.Println("Error loading production chain prices, try again.")
				continue
			}
			fmt.Println("Production chain prices updated.")

		case "R":
			region, ok := c.env.PromptRegion("Specify Region", "")
			if !ok {
				continue
			}
			if err := c.model.UpdateProductMarketPrices(p, region.RegionID); err != nil {
				c.logger.Errorf("unable to fetch market prices for region %d: %s", region.RegionID, err.Error())
				fmt.Println("Error loading production chain prices, try again.")
			}
			fmt.Printf("Updated %s target region to %s.\n", c.getProductName(p), c.getRegionName(p.MarketRegionID))

		case "?":
			fmt.Println()
			fmt.Println(text.WrapText(`The production chain editor is an interactive application for managing arbitrary production chains. Individual components can be tagged as either "buy" or "manufacture". Cost projections, with material efficiency and batch size accounted for, are updated accordingly. The target market region and target final sell price can also be modified for the production chain as a whole.

When invoking a tool and omitting an optional parameter, an interactive prompt will begin to collect the necessary information.

The current product is always line item 0, which can be used when specifying a line number.`, text.StandardTerminalWidthInChars))
			fmt.Println()
			colWidth := 7
			fmt.Printf("  %s Save the current production chain and exit the editor.\n", text.PadTextRight("S", colWidth))
			fmt.Printf("  %s View the production chain details.\n", text.PadTextRight("V", colWidth))
			fmt.Printf("  %s Update market prices.\n", text.PadTextRight("U", colWidth))
			fmt.Printf("  %s Set the market region for the production chain.\n", text.PadTextRight("R", colWidth))
			fmt.Printf("  %s Set the sell price per unit for the final product.\n", text.PadTextRight("P", colWidth))
			fmt.Printf("  %s Show detailed information for a specific chain item.\n", text.PadTextRight("D [#]", colWidth))
			fmt.Printf("  %s Set the production kind (buy or build) for a specific chain item.\n", text.PadTextRight("M [#]", colWidth))
			fmt.Printf("  %s Set the batch size for a specific chain item.\n", text.PadTextRight("B [#]", colWidth))
			fmt.Printf("  %s Set the material efficiency for a specific chain item.\n", text.PadTextRight("F [#]", colWidth))
			fmt.Printf("  %s Set the cost per unit for a specific chain item.\n", text.PadTextRight("C [#]", colWidth))

			fmt.Printf("  %s Quit the editor without saving changes.\n", text.PadTextRight("Q", colWidth))
			fmt.Printf("  %s Display this help text.\n", text.PadTextRight("?", colWidth))
			fmt.Println()
		}
	}
}
