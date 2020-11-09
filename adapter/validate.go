package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/layer5io/meshkit/smi"
)

type SmiTestOptions struct {
	Ctx         context.Context
	OpID        string
	Labels      map[string]string
	Annotations map[string]string
}

func (h *Adapter) ValidateSMIConformance(opts *SmiTestOptions) error {
	e := &Event{
		Operationid: opts.OpID,
		Summary:     status.Deploying,
		Details:     "None",
	}

	test, err := smi.New(opts.Ctx, opts.OpID, h.GetVersion(), strings.ToLower(h.GetName()), h.KubeClient)
	if err != nil {
		e.Summary = "Error while creating smi-conformance tool"
		e.Details = err.Error()
		h.StreamErr(e, ErrNewSmi(err))
		return err
	}

	result, err := test.Run(opts.Labels, opts.Annotations)
	if err != nil {
		e.Summary = fmt.Sprintf("Error while %s running smi-conformance test", result.Status)
		e.Details = err.Error()
		h.StreamErr(e, ErrRunSmi(err))
		return err
	}

	e.Summary = fmt.Sprintf("Smi conformance test %s successfully", result.Status)
	jsondata, _ := json.Marshal(result)
	e.Details = string(jsondata)
	h.StreamInfo(e)

	return nil
}
