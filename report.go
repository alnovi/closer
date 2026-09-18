package closer

import (
	"sync"
	"time"
)

type Report struct {
	Items []ReportItem
	mu    sync.Mutex
}

type ReportItem struct {
	Name     string
	Duration time.Duration
	Error    error
}

func NewReport() *Report {
	return &Report{}
}

func (r *Report) AddItem(name string, duration time.Duration, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Items = append(r.Items, ReportItem{
		Name:     name,
		Duration: duration,
		Error:    err,
	})
}

func (r *Report) HasErrors() bool {
	for _, i := range r.Items {
		if i.Error != nil {
			return true
		}
	}
	return false
}
