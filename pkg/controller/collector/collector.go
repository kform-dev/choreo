package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"google.golang.org/protobuf/proto"
)

type Collector interface {
	Start(ctx context.Context, once bool)
}

type TaskID struct {
	ReconcilerName string
	Group          string
	Kind           string
	Namespace      string
	Name           string
}

func (r TaskID) String() string {
	return fmt.Sprintf("%s.%s.%s.%s", r.ReconcilerName, r.Kind, r.Group, r.Name)
}

func New(
	reconcilerResultCh chan *runnerpb.Result,
	collectorResultCh chan *runnerpb.Once_Response,
) Collector {

	return &collector{
		reconcilerResultCh: reconcilerResultCh,
		collectorResultCh:  collectorResultCh,
		work:               map[TaskID]time.Time{},
		results:            []*runnerpb.Result{},
	}
}

type collector struct {
	reconcilerResultCh chan *runnerpb.Result
	collectorResultCh  chan *runnerpb.Once_Response
	m                  sync.Mutex
	work               map[TaskID]time.Time
	results            []*runnerpb.Result
	finishing          bool
	done               bool
	idle               int
	finish             time.Time
}

func (r *collector) Start(ctx context.Context, once bool) {
	log := log.FromContext(ctx)

	var waitTicker *time.Ticker
	var gracePeriod = 500 * time.Millisecond

	waitTicker = time.NewTicker(gracePeriod)
	defer func() {
		if waitTicker != nil {
			waitTicker.Stop()
			drainTicker(waitTicker)
		}
	}()

	var start time.Time
	running := true

	for running {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-r.reconcilerResultCh:
			if !ok {
				log.Debug("collector done...")
				return
			}
			if start.IsZero() {
				start = time.Now()
			}
			r.handleResult(ctx, result)
		case <-waitTicker.C:
			if once {
				if r.done {
					log.Debug("done", "elapsed time (sec)", r.finish.Sub(start).Seconds())
					r.collectorResultCh <- &runnerpb.Once_Response{
						Success:       true,
						ExecutionTime: fmt.Sprintf("%v", r.finish.Sub(start).Seconds()),
						Results:       r.results,
					}
				}
				if r.finishing {
					if !r.done {
						log.Debug("finishing", "elapsed time", r.finish.Sub(start).Seconds())
					}
					r.done = true
				}
				if r.idle == 3 {
					r.done = true
				}
				r.idle++
			}
		}
	}
}

func (r *collector) handleResult(ctx context.Context, result *runnerpb.Result) {
	r.m.Lock()
	defer r.m.Unlock()
	log := log.FromContext(ctx)
	cloneResult := proto.Clone(result).(*runnerpb.Result)
	r.results = append(r.results, cloneResult)
	taskID := TaskID{
		ReconcilerName: result.ReconcilerName,
		Group:          result.Resource.Group,
		Kind:           result.Resource.Kind,
		Namespace:      result.Resource.Namespace,
		Name:           result.Resource.Name,
	}
	log.Debug("collector result", "taskID", taskID.String(), "op", result.Operation.String(), "time", result.EventTime.AsTime().String())
	switch result.Operation {
	case runnerpb.Operation_ERROR:
		delete(r.work, taskID)
		log.Debug("execution failed", "taskID", taskID.String(), "error", result.Message)
		// context of the error
		r.collectorResultCh <- &runnerpb.Once_Response{
			Success: false,
			TaskId:  taskID.String(),
			Message: result.Message,
			Results: r.results,
		}
		return
	case runnerpb.Operation_REQUEUE:
		// this is a dummy time -> to ensure that the collector knows there is ongoing work
		r.work[taskID] = result.EventTime.AsTime()
		log.Debug("execution requeue", "taskID", taskID.String(), "error", result.Message)
	case runnerpb.Operation_START:
		r.work[taskID] = result.EventTime.AsTime()
	case runnerpb.Operation_STOP:
		delete(r.work, taskID)
		// this indicated the work queue is
		if len(r.work) == 0 {
			r.finish = time.Now()
			r.finishing = true
		}
	}
}

func drainTicker(ticker *time.Ticker) {
	select {
	case <-ticker.C:
	default:
	}
}
