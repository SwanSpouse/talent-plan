## map reduce 文档

### reduce 实现思路

1. 首先在map阶段，会从文件中读取输入数据，形成Key-value的形式，并将根据Key的Hash规则写到指定目录中。

2. 在reduce阶段，首先要确定reduce阶段的输入来自哪些文件。

3. 在确定好reduce 阶段的输入目录后，所有数据读入到内存中，在内存中形成key-> value list这样的数据结构。

4. 在reduce函数中，依次处理所有key->value list，并将结果写入到reduce的输出目录中。

5. 最后在所有的reduce任务接触之后，会将输入文件写入到notify chan中。


### url top 10 设计思路

![sort step](https://github.com/SwanSpouse/talent-plan/blob/master/tidb/mapreduce/images/urltop10.png?raw=true)

如上图所示，和urltop10的思路相似，只是在原有的两步mapreduce中间加入了一步。

1. map读取输入数据，随机分发到100个reduce方法中。

2. 这100个reduce方法每个reduce选择出自己的urltop10，最后再通过一步mapreduce将结果汇总，选择出全局的urltop10。


### 实验结果

``` json
Case0 PASS, dataSize=100MB, nMapFiles=5, cost=12.436368618s
Case1 PASS, dataSize=100MB, nMapFiles=5, cost=12.019567494s
Case2 PASS, dataSize=100MB, nMapFiles=5, cost=11.406856923s
Case3 PASS, dataSize=100MB, nMapFiles=5, cost=11.774048705s
Case4 PASS, dataSize=100MB, nMapFiles=5, cost=13.551784182s
Case5 PASS, dataSize=100MB, nMapFiles=5, cost=4.848235891s
Case6 PASS, dataSize=100MB, nMapFiles=5, cost=4.98017363s
Case7 PASS, dataSize=100MB, nMapFiles=5, cost=5.185854661s
Case8 PASS, dataSize=100MB, nMapFiles=5, cost=5.392663378s
Case9 PASS, dataSize=100MB, nMapFiles=5, cost=5.362078429s
Case10 PASS, dataSize=100MB, nMapFiles=5, cost=5.933029312s
Case0 PASS, dataSize=500MB, nMapFiles=10, cost=35.673652712s
Case1 PASS, dataSize=500MB, nMapFiles=10, cost=39.027565347s
Case2 PASS, dataSize=500MB, nMapFiles=10, cost=47.689750426s
Case3 PASS, dataSize=500MB, nMapFiles=10, cost=46.108344578s
panic: write /tmp/mr_homework/case4-500MB-10/mrtmp.Case4-Round0-5-12: no space left on device
```

限于个人电脑限制，没有能够跑完所有数据规模的测试。


### 性能瓶颈分析

#### CPU性能分析

```shell
(pprof) top20
Showing nodes accounting for 5.09mins, 76.90% of 6.62mins total
Dropped 405 nodes (cum <= 0.03mins)
Showing top 20 nodes out of 138
      flat  flat%   sum%        cum   cum%
  2.36mins 35.69% 35.69%   2.36mins 35.69%  runtime.pthread_cond_signal
  1.06mins 16.02% 51.72%   1.09mins 16.44%  syscall.syscall
  0.18mins  2.78% 54.50%   0.18mins  2.78%  cmpbody
  0.15mins  2.29% 56.79%   0.15mins  2.29%  runtime.memmove
  0.15mins  2.26% 59.05%   0.36mins  5.43%  runtime.scanobject
  0.13mins  2.03% 61.08%   0.17mins  2.63%  runtime.findObject
  0.12mins  1.80% 62.88%   0.12mins  1.81%  runtime.pthread_cond_wait
  0.11mins  1.70% 64.58%   0.30mins  4.61%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.TopN.func1
  0.11mins  1.64% 66.22%   0.16mins  2.35%  encoding/json.(*encodeState).string
  0.11mins  1.63% 67.85%   0.11mins  1.63%  runtime.memclrNoHeapPointers
  0.10mins  1.44% 69.29%   0.21mins  3.12%  runtime.mapassign_faststr
  0.09mins  1.38% 70.66%   0.35mins  5.25%  runtime.mallocgc
  0.07mins  1.04% 71.71%   0.07mins  1.04%  runtime.aeshashbody
  0.07mins     1% 72.71%   0.07mins     1%  runtime.usleep
  0.05mins  0.82% 73.52%   0.34mins  5.12%  encoding/json.structEncoder.encode
  0.05mins  0.79% 74.31%   0.45mins  6.80%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.genUniformCases.func1
  0.05mins  0.77% 75.08%   3.09mins 46.75%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.(*MRCluster).worker
  0.05mins  0.69% 75.77%   0.05mins  0.69%  hash/fnv.(*sum32a).Write
  0.04mins  0.58% 76.35%   0.04mins   0.6%  encoding/json.stateInString
  0.04mins  0.55% 76.90%   0.04mins  0.62%  runtime.spanOf
```

1. 从上述结果中可以看到，CPU有很长一部分时间是在等待状态：runtime.pthread_cond_signal。代码分析可知，在map reduce框架中，reduce的输入依赖于map的输出。reduce过程需要等
map过程完全结束才能够开始。这样就使得很大一部分时间浪费在等待map过程结束。此处是一个可以优化的点，因为一个map过程变慢，会拉低整体的效率，形成木桶效应。

2. 排序取TopN过程也是一个很耗时的过程，其实这里可以采用堆排序的思想对内存进行优化。但是时间复杂度是相同的。

3. 还有消耗时间较长的为：encode和decode过程、文件的写入过程。

#### memory 分析

```shell

(pprof) top20
Showing nodes accounting for 34.57GB, 99.65% of 34.70GB total
Dropped 31 nodes (cum <= 0.17GB)
Showing top 20 nodes out of 28
      flat  flat%   sum%        cum   cum%
   17.43GB 50.24% 50.24%    17.43GB 50.24%  bufio.NewWriterSize
    6.86GB 19.79% 70.02%    34.69GB   100%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.(*MRCluster).worker
    1.71GB  4.92% 74.94%     1.71GB  4.92%  encoding/json.(*decodeState).literalStore
    1.62GB  4.68% 79.62%     1.62GB  4.68%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.TopN
    1.49GB  4.30% 83.92%     1.67GB  4.80%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.ihash
    1.44GB  4.16% 88.07%     1.44GB  4.16%  bytes.makeSlice
    1.09GB  3.14% 91.21%     1.63GB  4.71%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.WordCountMap
    0.99GB  2.84% 94.05%     0.99GB  2.84%  strings.genSplit
    0.29GB  0.83% 94.89%     2.37GB  6.83%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.SelectCandidatesReduce
    0.27GB  0.78% 95.67%     0.59GB  1.70%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.MergeResultReduce
    0.27GB  0.78% 96.45%     0.27GB  0.78%  fmt.Sprintf
    0.27GB  0.77% 97.22%     1.71GB  4.93%  bytes.(*Buffer).grow
    0.21GB  0.62% 97.84%     0.21GB  0.62%  bytes.(*Buffer).String
    0.21GB   0.6% 98.44%     0.55GB  1.57%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.SelectCandidatesMap
    0.17GB   0.5% 98.94%     0.17GB   0.5%  hash/fnv.New32a
    0.14GB  0.42% 99.36%     0.22GB  0.63%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.MergeResultMap
    0.10GB   0.3% 99.65%     0.30GB  0.88%  github.com/SwanSpouse/talent-plan/tidb/mapreduce.WordCountReduce
         0     0% 99.65%     1.44GB  4.16%  bytes.(*Buffer).Grow
         0     0% 99.65%     0.27GB  0.77%  bytes.(*Buffer).Write
         0     0% 99.65%     1.74GB  5.01%  encoding/json.(*Decoder).Decode

```

1. 在代码中，内存消耗较高的两个流程依次是：WordCountMap、SelectCandidatesReduce；其中wordCount是需要将所有输入数据读入内存中；SelectCandidatesReduce则是因为需要进行排序。

2. 在上面CPU结果分析中已经提到，可以采用堆排序的思想：只在内存中保留Top10的数据，而不是全量加载再进行排序。以此来节省内存的空间。

