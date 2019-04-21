package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().Unix()))

// URLTop10 .
func URLTop10(nWorkers int) RoundsArgs {
	var args RoundsArgs
	// round 1: do url count
	args = append(args, RoundArgs{
		MapFunc:    WordCountMap,
		ReduceFunc: WordCountReduce,
		NReduce:    100,
	})

	// round 2: sort and get the 10 most frequent URLs for each reduce
	args = append(args, RoundArgs{
		MapFunc:    SelectCandidatesMap,
		ReduceFunc: SelectCandidatesReduce,
		NReduce:    10,
	})

	// round 3: merge all result and get final result
	args = append(args, RoundArgs{
		MapFunc:    MergeResultMap,
		ReduceFunc: MergeResultReduce,
		NReduce:    1,
	})
	return args
}

func WordCountMap(filename string, contents string) []KeyValue {
	lines := strings.Split(string(contents), "\n")
	kvs := make([]KeyValue, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		kvs = append(kvs, KeyValue{Key: l, Value: "1"})
	}
	return kvs
}

func WordCountReduce(key string, values []string) string {
	var sum int64
	for _, value := range values {
		count, _ := strconv.ParseInt(value, 10, 64)
		sum += count
	}
	return fmt.Sprintf("%s %d\n", key, sum)
}

func SelectCandidatesMap(filename string, contents string) []KeyValue {
	lines := strings.Split(string(contents), "\n")
	kvs := make([]KeyValue, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		kvs = append(kvs, KeyValue{Key: fmt.Sprintf("%d", ihash(l)), Value: l})
	}
	return kvs
}

func SelectCandidatesReduce(key string, values []string) string {
	cnts := make(map[string]int, len(values))
	for _, v := range values {
		v := strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		tmp := strings.Split(v, " ")
		n, err := strconv.Atoi(tmp[1])
		if err != nil {
			panic(err)
		}
		cnts[tmp[0]] = n
	}

	us, cs := TopN(cnts, 10)
	buf := new(bytes.Buffer)
	for i := range us {
		fmt.Fprintf(buf, "%s %d\n", us[i], cs[i])
	}
	return buf.String()
}

func MergeResultMap(filename string, contents string) []KeyValue {
	lines := strings.Split(string(contents), "\n")
	kvs := make([]KeyValue, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		kvs = append(kvs, KeyValue{Key: "", Value: l})
	}
	return kvs
}

func MergeResultReduce(key string, values []string) string {
	cnts := make(map[string]int, len(values))
	for _, v := range values {
		v := strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		tmp := strings.Split(v, " ")
		n, err := strconv.Atoi(tmp[1])
		if err != nil {
			panic(err)
		}
		cnts[tmp[0]] = n
	}

	us, cs := TopN(cnts, 10)
	buf := new(bytes.Buffer)
	for i := range us {
		fmt.Fprintf(buf, "%s: %d\n", us[i], cs[i])
	}
	return buf.String()
}
