/*
Copyright 2024 Nokia.

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

package resourceclient

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewClient(config *config.Config) (Client, error) {
	client := &client{
		config: config,
	}

	conn, err := grpc.NewClient(config.Address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(config.MaxMsgSize)),
	)
	if err != nil {
		return nil, err
	}
	client.client = resourcepb.NewResourceClient(conn)
	client.conn = conn
	return client, nil
}

type client struct {
	config *config.Config
	conn   *grpc.ClientConn
	client resourcepb.ResourceClient
}

func (r *client) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *client) Get(ctx context.Context, key types.NamespacedName, u runtime.Unstructured, opts ...GetOption) error {
	o := GetOptions{}
	o.ApplyOptions(opts)

	newu := &unstructured.Unstructured{
		Object: u.UnstructuredContent(),
	}
	newu.SetName(key.Name)
	newu.SetNamespace(key.Namespace)

	b, err := json.Marshal(newu)
	if err != nil {
		return err
	}
	rsp, err := r.client.Get(ctx, &resourcepb.Get_Request{
		Object: b,
		Options: &resourcepb.Get_Options{
			Branch:           o.Branch,
			Ref:              o.Ref,
			ShowManagedField: o.ShowManagedFields,
			Trace:            o.Trace,
			Origin:           o.Origin,
		},
	})
	if err != nil {
		return err
	}

	data := map[string]any{}
	if err := json.Unmarshal(rsp.Object, &data); err != nil {
		return err
	}
	u.SetUnstructuredContent(data)
	return nil
}

func (r *client) List(ctx context.Context, u runtime.Unstructured, opts ...ListOption) error {
	o := ListOptions{}
	o.ApplyOptions(opts)

	b, err := json.Marshal(u.UnstructuredContent())
	if err != nil {
		return err
	}
	rsp, err := r.client.List(ctx, &resourcepb.List_Request{
		Object: b,
		Options: &resourcepb.List_Options{
			ExprSelector:     o.ExprSelector,
			ShowManagedField: o.ShowManagedFields,
			Trace:            o.Trace,
			Origin:           o.Origin,
			Branch:           o.Branch,
			Ref:              o.Ref,
		},
	})
	if err != nil {
		return err
	}

	data := map[string]any{}
	if err := json.Unmarshal(rsp.Object, &data); err != nil {
		return err
	}
	u.SetUnstructuredContent(data)

	return nil
}

func (r *client) Apply(ctx context.Context, u runtime.Unstructured, opts ...ApplyOption) error {
	o := ApplyOptions{}
	o.ApplyOptions(opts)

	b, err := json.Marshal(u.UnstructuredContent())
	if err != nil {
		return err
	}
	rsp, err := r.client.Apply(ctx, &resourcepb.Apply_Request{
		Object: b,
		Options: &resourcepb.Apply_Options{
			Branch:       o.Branch,
			DryRun:       o.DryRun,
			FieldManager: o.FieldManager,
			Force:        o.Force,
			Trace:        o.Trace,
			Origin:       o.Origin,
		},
	})
	if err != nil {
		return err
	}

	data := map[string]any{}
	if err := json.Unmarshal(rsp.Object, &data); err != nil {
		return err
	}
	u.SetUnstructuredContent(data)

	return nil
}

func (r *client) Create(ctx context.Context, u runtime.Unstructured, opts ...CreateOption) error {
	o := CreateOptions{}
	o.ApplyOptions(opts)

	b, err := json.Marshal(u.UnstructuredContent())
	if err != nil {
		return err
	}
	rsp, err := r.client.Create(ctx, &resourcepb.Create_Request{
		Object: b,
		Options: &resourcepb.Create_Options{
			DryRun: o.DryRun,
			Branch: o.Branch,
			Trace:  o.Trace,
			Origin: o.Origin,
		},
	})
	if err != nil {
		return err
	}

	data := map[string]any{}
	if err := json.Unmarshal(rsp.Object, &data); err != nil {
		return err
	}
	u.SetUnstructuredContent(data)

	return nil
}

func (r *client) Update(ctx context.Context, u runtime.Unstructured, opts ...UpdateOption) error {
	o := UpdateOptions{}
	o.ApplyOptions(opts)

	b, err := json.Marshal(u.UnstructuredContent())
	if err != nil {
		return err
	}
	rsp, err := r.client.Update(ctx, &resourcepb.Update_Request{
		Object: b,
		Options: &resourcepb.Update_Options{
			DryRun: o.DryRun,
			Branch: o.Branch,
			Trace:  o.Trace,
			Origin: o.Origin,
		},
	})
	if err != nil {
		return err
	}

	data := map[string]any{}
	if err := json.Unmarshal(rsp.Object, &data); err != nil {
		return err
	}
	u.SetUnstructuredContent(data)

	return nil
}

func (r *client) Delete(ctx context.Context, u runtime.Unstructured, opts ...DeleteOption) error {
	o := DeleteOptions{}
	o.ApplyOptions(opts)

	b, err := json.Marshal(u.UnstructuredContent())
	if err != nil {
		return err
	}
	_, err = r.client.Delete(ctx, &resourcepb.Delete_Request{
		Object: b,
		Options: &resourcepb.Delete_Options{
			DryRun: o.DryRun,
			Branch: o.Branch,
			Trace:  o.Trace,
			Origin: o.Origin,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *client) Watch(ctx context.Context, u runtime.Unstructured, opts ...ListOption) chan *resourcepb.Watch_Response {
	o := ListOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	var stream resourcepb.Resource_WatchClient
	var err error

	b, err := json.Marshal(u.UnstructuredContent())
	if err != nil {
		log.Error("cannot unmarshal json", "err", err)
		return nil
	}

	rspCh := make(chan *resourcepb.Watch_Response)
	go func() {
		defer close(rspCh)
		for {
			select {
			case <-ctx.Done():
				// watch stoppped
				return
			default:
				if stream == nil {
					if stream, err = r.client.Watch(ctx, &resourcepb.Watch_Request{
						Object: b,
						Options: &resourcepb.Watch_Options{
							Branch:       o.Branch,
							Ref:          o.Ref,
							Watch:        o.Watch,
							ExprSelector: o.ExprSelector,
							Trace:        o.Trace,
							Origin:       o.Origin,
						},
					}); err != nil && !errors.Is(err, context.Canceled) {
						if statErr, ok := status.FromError(err); ok {
							switch statErr.Code() {
							case codes.Canceled:
								// dont log when context got cancelled
							default:
								log.Error("watch failed", "error", statErr.Err())
							}
						}
						time.Sleep(time.Second * 1) //- resilience for server crash
						// retry on failure
						continue
					}
				}
				rsp, err := stream.Recv()
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						if er, ok := status.FromError(err); ok {
							switch er.Code() {
							case codes.Canceled:
								log.Debug("resource client watch event recv error", "error", err.Error())
								// dont log when context got cancelled
							default:
								log.Error("failed to receive a message from stream", "error", err.Error())
							}
						}
						// clearing the stream will force the client to resubscribe in the next iteration
						stream = nil
						time.Sleep(time.Second * 1) //- resilience for server crash
						// retry on failure
						continue
					}
					if strings.Contains(err.Error(), "EOF") {
						log.Error("fail rcv", "error", err)
						continue
					}
				}
				log.Debug("resource client event received", "eventType", rsp.EventType.String())
				if rsp.EventType == resourcepb.Watch_ERROR {
					stream = nil
					time.Sleep(time.Second * 1) //- resilience for server error
					// retry on failure
					continue
				}
				rspCh <- rsp
			}
		}
	}()
	return rspCh
}
