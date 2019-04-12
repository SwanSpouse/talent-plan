package main

import (
	"sort"
)

func sortStep(retChan chan<- []int64, input []int64) {
	sort.Slice(input, func(i, j int) bool { return input[i] < input[j] })
	retChan <- input
}

func mergeStep(finalChan chan<- []int64, retChan <-chan []int64, jobCount int64) {
	var count int64 = 0

	ret := make([]int64, 0)
	for count < jobCount {
		select {
		case cur := <-retChan:
			temp := make([]int64, 0, len(ret)+len(cur))
			var i, j int
			for i < len(ret) && j < len(cur) {
				if ret[i] < cur[j] {
					temp = append(temp, ret[i])
					i++
				} else {
					temp = append(temp, cur[j])
					j++
				}
			}
			if i == len(ret) {
				temp = append(temp, cur[j:]...)
			} else if j == len(cur) {
				temp = append(temp, ret[i:]...)
			}
			ret = temp
			count += 1
		}
	}
	finalChan <- ret
}

// 归并排序
// MergeSort performs the merge sort algorithm.
// Please supplement this function to accomplish the home work.
func MergeSort(src []int64) {
	totalCount := len(src)
	batchCount := len(src)/7 + 1

	jobCount := totalCount / batchCount
	if totalCount%batchCount != 0 {
		jobCount += 1
	}
	retChan := make(chan []int64, jobCount)

	var startIndex = 0
	for startIndex < totalCount {
		endIndex := startIndex + batchCount
		if endIndex >= totalCount {
			endIndex = totalCount
		}
		// sort
		go sortStep(retChan, src[startIndex:endIndex])
		startIndex = endIndex
	}

	finalChan := make(chan []int64, 0)
	// merge result
	go mergeStep(finalChan, retChan, int64(jobCount))

	select {
	case ret := <-finalChan:
		copy(src, ret)
	}
	close(retChan)
	close(finalChan)
}
