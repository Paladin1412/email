package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"

	"../ds"
)

var kEmpty = ""

type docxParser struct {
	zipReader  *zip.Reader
	output     *ds.Array
	tableStack *ds.Stack
}

func (p *docxParser) ToHtml() (string, error) {
	for _, f := range p.zipReader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return kEmpty, err
			}
			defer rc.Close()
			err = p.docxXml2Text(rc, true)
			if err != nil {
				return kEmpty, err
			}
			return p.output.Join(""), nil
		}
	}

	return kEmpty, errors.New("Invalid docx format.")
}

// TODO <a:blip r:embed="rId8" cstate="print"/>
func (p *docxParser) docxXml2Text(input io.Reader, strict bool) error {
	output := p.output
	tableStack := p.tableStack

	x := xml.NewDecoder(input)
	x.Strict = strict

	for {
		d, err := x.Token()
		if d == nil || err != nil {
			break
		}

		switch v := d.(type) {
		case xml.CharData:
			output.Push(string(v))
		case xml.StartElement:
			if v.Name.Local == "tcPr" {
				p.handle_tcPr(x, &v)
			} else if v.Name.Local == "tr" {
				output.Push("<tr>\n")
				if !tableStack.IsEmpty() {
					// 0-based
					var layout = tableStack.Peek().(*ds.TableLayout)
					layout.IncRow()
				}
			} else if v.Name.Local == "tc" {
				output.Push(NewHtmlNode("td"))
				if !tableStack.IsEmpty() {
					// 0-based
					var layout = tableStack.Peek().(*ds.TableLayout)
					layout.IncCell(1, output.Length()-1)
				}
			} else if v.Name.Local == "tbl" {
				output.Push("<table>\n")
				tableStack.Push(ds.NewTableLayout())
			} else if v.Name.Local == "br" || v.Name.Local == "tab" {
				output.Push("\n")
			} else if v.Name.Local == "p" {
				output.Push("<p>")
			} else if v.Name.Local == "r" {
				// w:r
				output.Push(NewHtmlNode("span"))
			} else if v.Name.Local == "rPr" {
				p.handle_rPr(x, &v)
			}
		case xml.EndElement:
			if v.Name.Local == "tr" {
				output.Push("</tr>\n")
			} else if v.Name.Local == "tc" {
				output.Push("</td>")
			} else if v.Name.Local == "tbl" {
				tableStack.Pop()
				output.Push("</table>\n")
			} else if v.Name.Local == "p" {
				output.Push("</p>\n")
			} else if v.Name.Local == "r" {
				output.Push("</span>")
			}
		}
	}

	return nil
}

// 处理table cell的属性
func (p *docxParser) handle_tcPr(x *xml.Decoder, v *xml.StartElement) {
	output := p.output
	tableStack := p.tableStack

	var tcPr = &tcPrNode{}
	x.DecodeElement(tcPr, v)

	// 先找到当前的行的 tc 节点
	if tableStack.IsEmpty() {
		log.Println("tableStack should not be empty when meet tcPr")
		return
	}

	var layout = tableStack.Peek().(*ds.TableLayout)
	var rowIndex = layout.RowIndex
	var cellIndex = layout.CellIndex
	if rowIndex >= len(layout.Grids) {
		log.Printf("Invalid rowIndex = %d, len(layout.Grids) = %d\n",
			rowIndex, len(layout.Grids))
		return
	} else if cellIndex >= len(layout.Grids[rowIndex]) {
		log.Printf("Invalid cellIndex = %d, len(layout.Grids[%d]) = %d\n",
			cellIndex, rowIndex, len(layout.Grids[rowIndex]))
		return
	}

	var idx = layout.Grids[rowIndex][cellIndex]
	if idx >= output.Length() {
		log.Printf("Invalid array index = %d, array length is %d\n",
			idx, output.Length())
		return
	}

	// TODO getHtmlNode函数的角色很奇怪
	var n = getHtmlNode(output, idx)
	if n == nil {
		return
	}

	if tcPr.GridSpan != nil {
		n.Attr["colspan"] = tcPr.GridSpan.Value
		var gridSpan, _ = strconv.Atoi(tcPr.GridSpan.Value)
		if gridSpan > 1 {
			layout.IncCellAtRow(gridSpan-1, rowIndex, idx)
		}
	}

	if tcPr.VAlign != nil {
		n.Attr["valign"] = tcPr.VAlign.Value
	}

	if tcPr.Shd != nil {
		if tcPr.Shd.Color != "" {
			n.InlineStyles["color"] = colorValue(tcPr.Shd.Color)
		}
		if tcPr.Shd.Fill != "" {
			n.InlineStyles["background-color"] = colorValue(tcPr.Shd.Fill)
		}
	}

	if tcPr.VMerge != nil {
		if tcPr.VMerge.Value == "restart" {
			// 找到 output 里面 的 tc 节点，设置rowspan属性
			n.Attr["rowspan"] = "1"
		} else if tcPr.VMerge.Value == "" {
			// 找到本行 output 里面的 tc 节点，Dummy 设置为 true
			n.Dummy = true

			// 找到上一行 output 里面的 tc 节点，rowspan + 1
			// 上一行也可能是 Dummy 的节点，我们需要一直找下去，找到不是 Dummy 的节点为止
			rowIndex--
			for rowIndex >= 0 {
				if cellIndex >= len(layout.Grids[rowIndex]) {
					log.Printf("Invalid cellIndex = %d, len(layout.Grids[%d]) = %d\n",
						cellIndex, rowIndex, len(layout.Grids[rowIndex]))
					return
				}

				var idx = layout.Grids[rowIndex][cellIndex]
				// TODO getHtmlNode函数的角色很奇怪
				var n = getHtmlNode(output, idx)
				if n == nil {
					return
				}

				if !n.Dummy {
					var rowspan, _ = strconv.Atoi(n.Attr["rowspan"])
					n.Attr["rowspan"] = fmt.Sprintf("%d", rowspan+1)
					return
				}
				rowIndex--
			}
		}
	}
}

// 处理span的属性
func (p *docxParser) handle_rPr(x *xml.Decoder, v *xml.StartElement) {
	var rPr = &rPrNode{}
	x.DecodeElement(rPr, v)
	if p.output.Length() <= 0 {
		return
	}

	var last = p.output.Last()
	if r, ok := last.(*htmlNode); ok {
		if rPr.B != nil {
			r.InlineStyles["font-weight"] = "bold"
		}
		if rPr.I != nil {
			r.InlineStyles["font-style"] = "italic"
		}
		if rPr.U != nil {
			r.InlineStyles["text-decoration"] = "underline"
		}

		// TODO 颜色值有的是 black, white 之类的，需要特殊对待一下
		if rPr.Color != nil && rPr.Color.Value != "" {
			r.InlineStyles["color"] = colorValue(rPr.Color.Value)
		}
		if rPr.Highlight != nil && rPr.Highlight.Value != "" {
			r.InlineStyles["background-color"] = colorValue(rPr.Highlight.Value)
		}

		// 字体和字号先不处理了，不然效果看起来不是很好看
		if rPr.Sz != nil && rPr.Sz.Value != "" {
			// r.InlineStyles["font-size"] = rPr.Sz.Value + "px"
		}
		if rPr.RFonts != nil && rPr.RFonts.Ascii != "" {
			// r.InlineStyles["font-family"] = rPr.RFonts.Ascii
		}
	}
}

func NewDocxParser(r *zip.Reader) *docxParser {
	return &docxParser{
		zipReader:  r,
		output:     ds.NewArray(),
		tableStack: ds.NewStack(),
	}
}

// docx 2 html
func DOCX2Html(file string) (string, error) {
	inputs, err := ioutil.ReadFile(file)
	if err != nil {
		return kEmpty, err
	}

	r, err := zip.NewReader(bytes.NewReader(inputs), int64(len(inputs)))
	if err != nil {
		return kEmpty, err
	}

	return NewDocxParser(r).ToHtml()
}