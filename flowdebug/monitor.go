package flowdebug

type chanInterface interface {
	Cap() int
	Len() int
}

type chanMonitor struct {
	ch     chanInterface
	val    float32
	metric Metric
}

func newChanMonitor(ch chanInterface, m Metric) *chanMonitor {
	if ch.Cap() == 0 {
		return nil
	}
	return &chanMonitor{
		ch:     ch,
		metric: m,
	}
}

func (m *chanMonitor) Sample(weight float32) {
	if m == nil {
		return
	}
	cnt := float32(m.ch.Len())
	m.val = m.val*(1-weight) + cnt*weight
}

func (m *chanMonitor) Record() {
	if m == nil {
		return
	}
	m.metric.Store(int32(m.val))
}
