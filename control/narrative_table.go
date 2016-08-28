package control

import (
	"errors"
	"regexp"

	"github.com/jbowtie/gokogiri/xml"
	"github.com/opencontrol/fedramp-templater/opencontrols"
)

func findSectionKey(row xml.Node) (section string, err error) {
	re := regexp.MustCompile(`Part ([a-z])`)
	subMatches := re.FindSubmatch([]byte(row.Content()))
	if len(subMatches) != 2 {
		err = errors.New("No Parts found.")
		return
	}
	section = string(subMatches[1])
	return
}

func fillRow(row xml.Node, data opencontrols.Data, control string, section string) (err error) {
	paragraphNodes, err := row.Search(`./w:tc[last()]/w:p[1]`)
	if err != nil {
		return
	}
	paragraphNode := paragraphNodes[0]

	err = paragraphNode.SetChildren(`<w:r><w:t></w:t></w:r>`)
	if err != nil {
		return
	}
	textCell := paragraphNode.FirstChild().FirstChild()

	narrative := data.GetNarrative(control, section)
	textCell.SetContent(narrative)
	return
}

type NarrativeTable struct {
	tbl table
}

func NewNarrativeTable(root xml.Node) NarrativeTable {
	tbl := table{Root: root}
	return NarrativeTable{tbl}
}

func (t *NarrativeTable) SectionRows() ([]xml.Node, error) {
	// skip the header row
	return t.tbl.searchSubtree(`.//w:tr[position() > 1]`)
}

func (t *NarrativeTable) Fill(openControlData opencontrols.Data) (err error) {
	control, err := t.tbl.controlName()
	if err != nil {
		return
	}

	rows, err := t.SectionRows()
	if err != nil {
		return
	}

	if len(rows) == 1 {
		// singular narrative
		row := rows[0]
		fillRow(row, openControlData, control, "")
	} else {
		// multiple parts
		for _, row := range rows {
			sectionKey, err := findSectionKey(row)
			if err != nil {
				return err
			}

			err = fillRow(row, openControlData, control, sectionKey)
			if err != nil {
				return err
			}
		}
	}

	return
}
