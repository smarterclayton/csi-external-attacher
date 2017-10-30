/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"github.com/golang/glog"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

// trivialHandler is a handler that marks all VolumeAttachments as attached.
// It's used for CSI drivers that don't support ControllerPulishVolume call.
// It uses no finalizer, deletion of VolumeAttachment is instant (as there is
// nothing to detach).
type trivialHandler struct {
	client kubernetes.Interface
}

var _ Handler = &trivialHandler{}

func NewTrivialHandler(client kubernetes.Interface) Handler {
	return &trivialHandler{client}
}
func (h *trivialHandler) SyncNewOrUpdatedVolumeAttachment(va *storagev1.VolumeAttachment, queue workqueue.RateLimitingInterface) {
	glog.V(4).Infof("Trivial sync[%s] started", va.Name)
	if !va.Status.Attached {
		// mark as attached
		if err := h.markAsAttached(va); err != nil {
			glog.Warningf("Error saving VolumeAttachment %s as attached: %s", va.Name, err)
			queue.AddRateLimited(va.Name)
			return
		}
		glog.V(2).Infof("Marked VolumeAttachment %s as attached", va.Name)
	}
	queue.Forget(va.Name)
}

func (h *trivialHandler) markAsAttached(va *storagev1.VolumeAttachment) error {
	clone := va.DeepCopy()
	clone.Status.Attached = true
	_, err := h.client.StorageV1().VolumeAttachments().Update(clone)
	return err
}