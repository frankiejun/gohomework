# 利用中间件优化web模块的日志打印

在user的handler中增加2行代码  
var err error   
ctx.Set("error", &err)

增加中间件处理方法：  
```golang

func logingHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
    c.Next()
    errptr, exists := c.Get("error")
    if exists && errptr != nil {
        if err, ok := errptr.(*error); ok {
            zap.L().Debug("处理请求失败:", zap.Error(*err))
        }
    }
  }
}
```  
运行效果  
![](https://cdn.jsdelivr.net/gh/frankiejun/pics@main/img/zap.png)