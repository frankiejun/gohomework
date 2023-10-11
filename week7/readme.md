# 利用中间件优化web模块的日志打印

在user的handler中增加2行代码  
var err error   
ctx.Set("error", &err)

增加中间件处理方法：  
```golang

func logingHandler() gin.HandlerFunc {  
	return func(c *gin.Context) {  
        c.Next()
        anyerr, exists := c.Get("error")  
         if exists && anyerr != nil {  
            if errptr, ok := anyerr.(*error); ok {  
                 if *errptr != nil {  
                 zap.L().Error("处理请求失败:", zap.Error(*errptr))  
                 }  
             }  
         }  
	}  
}  
```  
运行效果  
![](https://cdn.jsdelivr.net/gh/frankiejun/pics@main/img/zap_error.png)