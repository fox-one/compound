package rest

import (
	"compound/core"
	"compound/handler/param"
	"compound/handler/render"
	"compound/handler/views"
	"errors"
	"fmt"
	"net/http"
)

func handleProposals(proposals core.ProposalStore, proposalz core.ProposalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var params struct {
			Cursor int64 `json:"cursor"`
			Offset int64 `json:"offset"`
		}
		if e := param.Binding(r, &params); e != nil {
			render.BadRequest(w, e)
			return
		}

		if params.Offset > 0 {
			params.Cursor = params.Offset
		}

		const LIMIT = 50
		proposals, err := proposals.List(ctx, params.Cursor, LIMIT)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		pviews := views.ProposalViews(proposals)
		var nextCursor string
		if len(proposals) == LIMIT {
			nextCursor = fmt.Sprint(proposals[LIMIT-1].ID)
		}

		render.JSON(w, render.H{
			"data": render.H{
				"proposals": pviews,
				"pagination": render.H{
					"next_cursor": nextCursor,
					"has_next":    nextCursor != "",
				},
			},
		})
	}
}

func handleProposal(proposals core.ProposalStore, proposalz core.ProposalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		trace := param.String(r, "trace_id")
		proposal, found, err := proposals.Find(ctx, trace)
		if err != nil {
			if found {
				render.BadRequest(w, errors.New("proposal not found"))
				return
			}

			render.BadRequest(w, err)
			return
		}

		view := views.ProposalView(*proposal)
		pitems, err := proposalz.ListItems(ctx, proposal)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		view.Items = views.ProposalItemViews(pitems)

		render.JSON(w, render.H{
			"data": view,
		})
	}
}
