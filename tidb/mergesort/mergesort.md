## merge sort 文档

### 实现思路

1. 首先会将src数组的数据进行拆分，启动N个goroutine，每个goroutine处理n条数据（最后一个goroutine可能不足n条）。

2. sort阶段：每个goroutine的工作很简单，会将获取到的数据进行排序，并发送到chan中。

```golang

func sortStep(retChan chan<- []int64, input []int64) {
	sort.Slice(input, func(i, j int) bool { return input[i] < input[j] })
	retChan <- input
}

```

3. merge阶段：不断的从chan获取已经排好序的数据片段。并将片段依次合并成有序的序列。

### 结果比对

![结果](https://raw.githubusercontent.com/SwanSpouse/talent-plan/master/tidb/mergesort/images/make%20bench.png)

通过上述结果可以看到，并发的mergeSort要由于normalMergeSort

### 性能瓶颈分析

#### goroutine 个数

根据实验思路所描述，会将输入数据拆分到N个goroutine中依次进行排序。这里N的选择，对性能的影响较大。

```json
N:20	 	 BenchmarkMergeSort-4		   1	1873263294 ns/op 	
N:10	 	 BenchmarkMergeSort-4		   1	1631306878 ns/op  
N:5		   BenchmarkMergeSort-4		   1	1501145824 ns/op  
N:8		   BenchmarkMergeSort-4		   1	1580304375 ns/op  
N:7		   BenchmarkMergeSort-4		   1	1485027156 ns/op  
N:6		   BenchmarkMergeSort-4		   1	1497006597 ns/op  
N:1		   BenchmarkNormalSort-4	 	 1	3113600083 ns/op
```

从上述数据中可以看出，并不是并发数量越高，处理的效率越快。这个之间存在着一种动态平衡。经过试验，当goroutine的个数为7的时候，在当前机器上的性能表现效果最好。


#### pprof结果分析

![pprof](https://github.com/SwanSpouse/talent-plan/blob/master/tidb/mergesort/images/pprof-top.png?raw=true)

![sort step](https://github.com/SwanSpouse/talent-plan/blob/master/tidb/mergesort/images/pprof%20list%20sort%20step.png?raw=true)

从上面两图中可以看出，最耗时的两个操作为:
* sort阶段,每个goroutine对部分src进行排序。
* merge阶段, 由于merge阶段要等待所有的goroutine运行结果，并且对结果进行合并。导致耗时较长。

针对上面两个比较耗时的操作，还没有想到比较好的优化方案。


#### 预估slice的cap，提高性能

![slice cap](https://github.com/SwanSpouse/talent-plan/blob/master/tidb/mergesort/images/list%20merge%20step.png?raw=true)

从上图中可以看到，最耗时的是slice的append操作。由于我们可以提前知道两个slice merge之后的总长度，所以在创建slice的时候就会指定cap，避免在append的过程中slice进行扩容。
