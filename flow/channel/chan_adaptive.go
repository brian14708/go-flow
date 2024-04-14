package channel

import (
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/VividCortex/ewma"

	"github.com/brian14708/go-flow/flowtype"
)

type adaptiveChan struct {
	sendEnd reflect.Value
	buf     chan interface{}
	recvEnd reflect.Value

	*options
}

func newAdaptiveChan(src, dst reflect.Type, opt *options) (Channel, error) {
	return &adaptiveChan{
		sendEnd: reflect.MakeChan(reflect.ChanOf(reflect.BothDir, src), 1),
		buf:     make(chan interface{}, opt.size-1),
		recvEnd: reflect.MakeChan(reflect.ChanOf(reflect.BothDir, dst), 0),
		options: opt,
	}, nil
}

func (a *adaptiveChan) Drain() {
	drain(a.recvEnd, a.drainRate, a.dropHandler)
}

func (a *adaptiveChan) Close() {
	a.sendEnd.Close()
}

// adaptive number of inflight elements to a multiple of bandwidth delay product,
// similar to "BBR: Congestion-Based Congestion Control".
//
// - estimiate processing delay with send-block time.
func (l *adaptiveChan) Serve() {
	recv := flowtype.ChanRecver(l.sendEnd.Interface())

	var (
		mu       sync.Mutex
		cond     = sync.NewCond(&mu)
		inflight int
		wnd      = 1
	)

	go func() {
		for {
			v, ok := recv(nil, true)
			if !ok {
				close(l.buf)
				return
			}

			mu.Lock()
			for inflight >= wnd {
				cond.Wait()
			}
			inflight++
			mu.Unlock()

			l.buf <- v
		}
	}()

	var (
		bdp            = ewma.NewMovingAverage(5)
		prevUpdate     time.Time
		totalCnt       int
		slowCnt        int
		processTime    time.Duration
		maxProcessTime time.Duration
	)

	send := flowtype.ChanSender(l.recvEnd.Interface())
	for v := range l.buf {
		for _, interceptor := range l.interceptor {
			v = interceptor(v)
		}

		beginSend := time.Now()
		send(v, nil, true)
		endSend := time.Now()

		singleProcessTime := endSend.Sub(beginSend)
		if singleProcessTime >= maxProcessTime {
			slowCnt++
			maxProcessTime = singleProcessTime
		} else {
			maxProcessTime = maxProcessTime * time.Duration(98) / time.Duration(100)
		}
		processTime += singleProcessTime

		dur := endSend.Sub(prevUpdate)
		updateWnd := 0.0
		totalCnt++
		if slowCnt >= 2 || (dur >= time.Second && totalCnt >= 16) {
			bdp.Add(
				(float64(totalCnt) / float64(dur)) * float64(processTime),
			)
			totalCnt, processTime, slowCnt = 0, 0, 0
			prevUpdate = endSend
			if bdp.Value() != 0 {
				updateWnd = bdp.Value()
			} else {
				updateWnd = -1
			}
		}

		mu.Lock()
		if updateWnd > 0 {
			wnd = int(math.Round(l.adaptiveGain * updateWnd))
		} else if updateWnd == -1 {
			wnd = int(math.Round(float64(wnd) * (1 + l.adaptiveGain)))
		}
		if wnd <= 0 {
			wnd = 1
		}
		inflight--
		mu.Unlock()
		cond.Signal()
	}
	l.recvEnd.Close()
}

func (l *adaptiveChan) Cap() int { return cap(l.buf) + 1 }
func (l *adaptiveChan) Len() int { return len(l.buf) + l.sendEnd.Len() }

func (l *adaptiveChan) DropMessage(v interface{}) {
	if l.dropHandler != nil {
		l.dropHandler(v)
	}
}

func (l *adaptiveChan) AssignTo(d reflect.ChanDir, p interface{}) {
	el := reflect.ValueOf(p).Elem()
	switch d {
	case reflect.RecvDir:
		el.Set(l.recvEnd)
	case reflect.SendDir:
		el.Set(l.sendEnd)
	default:
		panic("invalid chan direction")
	}
}
