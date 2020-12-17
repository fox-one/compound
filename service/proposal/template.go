package proposal

import (
	"compound/core"

	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

func codeBlock(data []byte, tag string) []byte {
	var b bytes.Buffer
	_, _ = fmt.Fprintf(&b, "```%s\n", tag)
	b.Write(data)
	b.WriteByte('\n')
	b.WriteString("```")

	return b.Bytes()
}

const proposalTpl = `### #{{.ID}} NEW PROPOSAL "{{.Action}}"

{{.Proposal}}
`

type view struct {
	ID       int64
	Action   string
	Proposal string
}

func renderProposal(p *core.Proposal) []byte {
	v := view{
		ID:     p.ID,
		Action: p.Action.String(),
	}

	data, _ := json.MarshalIndent(p, "", "  ")
	v.Proposal = string(codeBlock(data, "json"))

	t, err := template.New("-").Parse(proposalTpl)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, v); err != nil {
		panic(err)
	}

	return b.Bytes()
}

const approvedByTpl = `✅ Approved By {{.Reviewer}}

({{.ApprovedCount}} Votes In Total)
`

func renderApprovedBy(p *core.Proposal, member *core.Member) []byte {
	t, err := template.New("-").Parse(approvedByTpl)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, map[string]interface{}{
		"ApprovedCount": len(p.Votes),
		"Reviewer":      member.ClientID,
	}); err != nil {
		panic(err)
	}

	return b.Bytes()
}

const passedTpl = "🎉 Proposal Passed"
