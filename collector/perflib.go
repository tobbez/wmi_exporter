package collector

import (
	"fmt"
	"reflect"

	"github.com/leoluk/perflib_exporter/perflib"
)

const (
	perfCounterRawcountHex         uint32 = 0
	perfCounterLargeRawcountHex           = 256
	perfCounterText                       = 2816
	perfCounterRawcount                   = 65536
	perfCounterLargeRawcount              = 65792
	perfDoubleRaw                         = 73728
	perfCounterDelta                      = 4195328
	perfCounterLargeDelta                 = 4195584
	perfSampleCounter                     = 4260864
	perfCounterQueuelenType               = 4523008
	perfCounterLargeQueuelenType          = 4523264
	perfCounter100nsQueuelenType          = 5571840
	perfCounterObjTimeQueuelenType        = 6620416
	perfCounterCounter                    = 272696320
	perfCounterBulkCount                  = 272696576
	perfRawFraction                       = 537003008
	perfCounterTimer                      = 541132032
	perfPrecisionSystemTimer              = 541525248
	perf100nsecTimer                      = 542180608
	perfPrecision100nsTimer               = 542573824
	perfObjTimeTimer                      = 543229184
	perfPrecisionObjectTimer              = 543622400
	perfSampleFraction                    = 549585920
	perfCounterTimerInv                   = 557909248
	perf100nsecTimerInv                   = 558957824
	perfCounterMultiTimer                 = 574686464
	perf100nsecMultiTimer                 = 575735040
	perfCounterMultiTimerInv              = 591463680
	perf100nsecMultiTimerInv              = 592512256
	perfAverageTimer                      = 805438464
	perfElapsedTime                       = 807666944
	perfCounterNodata                     = 1073742336
	perfAverageBulk                       = 1073874176
	perfSampleBase                        = 1073939457
	perfAverageBase                       = 1073939458
	perfRawBase                           = 1073939459
	perfPrecisionTimestamp                = 1073939712
	perfLargeRawBase                      = 1073939715
	perfCounterMultiBase                  = 1107494144
	perfCounterHistogramType              = 2147483648
)

func UnmarshalObject(obj *perflib.PerfObject, vs interface{}) error {
	if obj == nil {
		return fmt.Errorf("obj is nil")
	}
	rv := reflect.ValueOf(vs)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%v is nil or not a pointer to slice", reflect.TypeOf(vs))
	}
	ev := rv.Elem()
	if ev.Kind() != reflect.Slice {
		return fmt.Errorf("%v is not slice", reflect.TypeOf(vs))
	}

	// Ensure sufficient length
	if ev.Cap() < len(obj.Instances) {
		nvs := reflect.MakeSlice(ev.Type(), len(obj.Instances), len(obj.Instances))
		ev.Set(nvs)
	}

	for idx, instance := range obj.Instances {
		target := ev.Index(idx)
		rt := target.Type()

		counters := make(map[string]*perflib.PerfCounter, len(instance.Counters))
		for _, ctr := range instance.Counters {
			counters[ctr.Def.Name] = ctr
		}

		for i := 0; i < target.NumField(); i++ {
			f := rt.Field(i)
			tag := f.Tag.Get("perflib")
			if tag == "" {
				continue
			}

			ctr, found := counters[tag]
			if !found {
				return fmt.Errorf("could not find counter %q on instance", tag)
			}
			if !target.Field(i).CanSet() {
				return fmt.Errorf("tagged field %v cannot be written to", f)
			}

			switch ctr.Def.CounterType {
			case perfElapsedTime:
				target.Field(i).SetFloat(float64(ctr.Value-116444736000000000) / float64(obj.Frequency))
			case perf100nsecTimer:
				target.Field(i).SetFloat(float64(ctr.Value) * ticksToSecondsScaleFactor)
			default:
				fmt.Println(ctr.Def.CounterType, ctr.Def.Name)
				target.Field(i).SetFloat(float64(ctr.Value))
			}
		}

		if instance.Name != "" && target.FieldByName("Name").CanSet() {
			target.FieldByName("Name").SetString(instance.Name)
		}
	}

	return nil
}
