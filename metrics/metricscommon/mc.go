package metricscommon

import (
	"fmt"
	"time"

	"github.com/myfantasy/mfctx"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCommon struct {
	Start           *prometheus.CounterVec
	FinishTotal     *prometheus.CounterVec
	FinishTimeHist  *prometheus.HistogramVec
	FinishTimeTotal *prometheus.CounterVec
	Alarm           map[string]map[string]bool
	AlarmSegment    map[string]bool
}

var _ mfctx.MetricsProvider = &MetricsCommon{}

func NewMetricsCommon() *MetricsCommon {
	constLabels := prometheus.Labels{
		"version":  mfctx.GetAppVersion(),
		"app_name": mfctx.GetAppName(),
		"app_id":   mfctx.GetAppID(),
	}

	return &MetricsCommon{
		Start: prometheus.NewCounterVec(
			// nolint:promlinter
			prometheus.CounterOpts{
				Name:        "method_run_start",
				Help:        "Total amount of runnings",
				ConstLabels: constLabels,
			}, []string{"segment", "method"},
		),
		FinishTotal: prometheus.NewCounterVec(
			// nolint:promlinter
			prometheus.CounterOpts{
				Name:        "method_run_finish",
				Help:        "How many HTTP runnings finish, partitioned",
				ConstLabels: constLabels,
			}, []string{
				"segment", "method", "status_code", "alarm",
			},
		),
		FinishTimeHist: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "method_run_finish_hist",
				Help:        "Total amount of time spent on the runnings finish",
				ConstLabels: constLabels,
				Buckets:     []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000},
			}, []string{
				"segment", "method", "status_code", "alarm",
			},
		),
		FinishTimeTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "method_run_finish_time_total",
				Help:        "Total amount of time spent on the runnings finish",
				ConstLabels: constLabels,
			}, []string{
				"segment", "method", "status_code", "alarm",
			},
		),
	}
}

func (mc *MetricsCommon) AutoRegister() *MetricsCommon {
	mc.MustRegister(prometheus.DefaultRegisterer)

	return mc
}

func (mc *MetricsCommon) MustRegister(registerer prometheus.Registerer) *MetricsCommon {
	registerer.MustRegister(
		mc.Start,
		mc.FinishTotal,
		mc.FinishTimeHist,
		mc.FinishTimeTotal,
	)

	return mc
}

func (mc *MetricsCommon) WriteMetricRequest(c *mfctx.Crumps, segment, method string) {
	if mc == nil {
		return
	}

	mc.Start.WithLabelValues(segment, method).Inc()
}
func (mc *MetricsCommon) WriteMetricResponse(c *mfctx.Crumps, mRequest time.Time, segment, method string, statusCode string) {
	if mc == nil {
		return
	}

	var alarm bool
	if statusCode == mfctx.MsgError {
		asg := mc.AlarmSegment[segment]
		msg := mc.Alarm[segment][method]

		alarm = asg || msg
	}

	responseLabels := []string{segment, method, statusCode, fmt.Sprint(alarm)}
	diff := time.Since(mRequest).Milliseconds()

	mc.FinishTotal.WithLabelValues(responseLabels...).Inc()
	mc.FinishTimeHist.WithLabelValues(responseLabels...).Observe(float64(diff))
	mc.FinishTimeTotal.WithLabelValues(responseLabels...).Add(float64(diff))
}
