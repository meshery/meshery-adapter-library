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

import "github.com/layer5io/meshery-adapter-library/meshes"

func (h *Adapter) StreamErr(e *meshes.EventsResponse, err error) {
	h.Log.Error(err)
	e.EventType = 2
	//Putting this under a go routine so that this function is never blocking. If this push is performed synchronously then the call will be blocking in case
	//when the channel is full with no client to receive the events. This blocking may cause many operations to not return.
	go func() {
		h.EventStreamer.Publish(e)
		h.Log.Info("Event stored and sent successfully")
	}()
}

func (h *Adapter) StreamInfo(e *meshes.EventsResponse) {
	h.Log.Info("Sending event")
	e.EventType = 0
	//Putting this under a go routine so that this function is never blocking. If this push is performed synchronously then the call will be blocking in case
	//when the channel is full with no client to receive the events. This blocking may cause many operations to not return.
	go func() {
		h.EventStreamer.Publish(e)
		h.Log.Info("Event stored and sent successfully")
	}()
}
