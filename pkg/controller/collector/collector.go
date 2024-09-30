package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/henderiw/logger/log"
	reconcileresult "github.com/kform-dev/choreo/pkg/controller/collector/result"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
)

type Collector interface {
	Start(ctx context.Context, once bool)
}

func New(
	reconcilerResultCh chan reconcileresult.Result,
	collectorResultCh chan *runnerpb.Once_Response,
) Collector {

	return &collector{
		reconcilerResultCh: reconcilerResultCh,
		collectorResultCh:  collectorResultCh,
		work:               map[reconcileresult.ReconcileRef]time.Time{},
		results:            map[string]*runnerpb.Once_Operations{},
	}
}

type collector struct {
	reconcilerResultCh chan reconcileresult.Result
	collectorResultCh  chan *runnerpb.Once_Response
	work               map[reconcileresult.ReconcileRef]time.Time
	results            map[string]*runnerpb.Once_Operations
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

func (r *collector) handleResult(ctx context.Context, result reconcileresult.Result) {
	log := log.FromContext(ctx)
	ref := result.ReconcileRef
	log.Debug("collector result", "ref", ref.String(), "op", result.Operation.String(), "time", result.Time)

	if _, ok := r.results[ref.ReconcilerName]; !ok {
		r.results[ref.ReconcilerName] = &runnerpb.Once_Operations{
			OperationCounts: map[string]int32{},
		}
	}
	r.results[ref.ReconcilerName].OperationCounts[result.Operation.String()]++
	r.processOperation(ctx, result)
}

func (r *collector) processOperation(ctx context.Context, result reconcileresult.Result) {
	log := log.FromContext(ctx)
	ref := result.ReconcileRef
	switch result.Operation {
	case runnerpb.Operation_ERROR:
		startTime := r.getStartTime(ref)
		if startTime != nil {
			result.Elapsed = result.Time.Sub(*startTime)
		}
		delete(r.work, result.ReconcileRef)
		//r.reconcileResults = append(r.reconcileResults, result)
		log.Debug("execution failed", "ref", ref.String(), "error", result.Message)
		// context of the error
		r.collectorResultCh <- &runnerpb.Once_Response{
			Success:      false,
			ReconcileRef: ref.String(),
			Message:      result.Message,
		}
		return
	case runnerpb.Operation_REQUEUE:
		startTime := r.getStartTime(ref)
		if startTime != nil {
			result.Elapsed = result.Time.Sub(*startTime)
		}
		//r.reconcileResults = append(r.reconcileResults, result)
		r.work[ref] = result.Time // this is a dummy time
		log.Debug("execution requeue", "ref", ref.String(), "error", result.Message)
	case runnerpb.Operation_START:
		r.work[ref] = result.Time
	case runnerpb.Operation_STOP:
		startTime := r.getStartTime(ref)
		if startTime != nil {
			result.Elapsed = result.Time.Sub(*startTime)
		}
		//r.reconcileResults = append(r.reconcileResults, result)
		delete(r.work, result.ReconcileRef)
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

func (r *collector) getStartTime(ref reconcileresult.ReconcileRef) *time.Time {
	reconcileStartTime, exists := r.work[ref]
	if !exists {
		return nil
	}
	return &reconcileStartTime
}
