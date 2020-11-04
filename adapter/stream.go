// Copyright 2020 Layer5, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapter

import (
	"github.com/layer5io/gokit/errors"
)

type Event struct {
	Operationid string `json:"operationid,omitempty"`
	EType       int32  `json:"type,string,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Details     string `json:"details,omitempty"`
}

func (h *Adapter) StreamErr(e *Event, err error) {
	h.Log.Err(errors.GetCode(err), err.Error())
	e.EType = 2
	*h.Channel <- e
}

func (h *Adapter) StreamInfo(e *Event) {
	h.Log.Info("Sending event")
	e.EType = 0
	*h.Channel <- e
}
