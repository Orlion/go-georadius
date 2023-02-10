# go-georadius

redis `georadius`命令会在redis server中进行距离计算，而redis主线程又是单线程的，无法利用多核增加这部分计算的性能。

因此可以将计算部分提升到redis client中进行，仅将redis server作为数据库，实现”存储-计算“分离。

可以利用这个库将原本的`georadius`命令转为多个`zrangebyscore`命令，粗筛出坐标点后再在go程序中进行利用该库进行距离计算，筛选出不符合距离条件的坐标点。