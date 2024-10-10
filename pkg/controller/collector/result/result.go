package result

import (
	"fmt"
	"time"

	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type Result struct {
	Operation    choreopb.Operation
	ReconcileID  types.UID
	ReconcileRef ReconcileRef
	Message      string
	Time         time.Time
	Elapsed      time.Duration
}

type ReconcileRef struct {
	ReconcilerName string
	GVK            schema.GroupVersionKind
	Req            types.NamespacedName
}

func (r ReconcileRef) String() string {
	return fmt.Sprintf("%s.%s.%s.%s", r.ReconcilerName, r.GVK.Kind, r.GVK.Group, r.Req.Name)
}
