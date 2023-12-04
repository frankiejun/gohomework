

1.修改文件见：pkg->saramax->consumer_handler_func.go, ioc->gin.go   

2.用了counter指标监控sarama处理消息失败的记录  

3.设置监控，对处理消息失败的记录进行监控，以便及时处理异常情况  

4.有个疑问，用counter的inc方法可以进行监控告警，但告警处理完后，这个counter没法清0，那告警如何恢复？  
