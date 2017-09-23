This proxy is between short connection  and long connection.

这个Demo项目用来演示短连接如何变成长连接。
一种常见的情景，进来的是四面八方客户的http短连接，比如呼叫信息。
但是后台的服务器为了避免频繁建立连接可能的开销，实现的是一个长连接，
这时候一个代理就需要了，实现的是短连接到长连接的转换。

项目里的RestServer实现的是提供长连接的服务器。

RestProxy实现的是短连接到长连接转换的代理服务器.它是基于beego框架实现的。beego是国人开发的golang http服务框架，有非常完善的文档和教程。

RestClient实现的是发起短连接的客户端。

启动顺序是
RestServer->RestProxy->RestClient 

通过比较，该项目利用了golang支持高并发的特性，性能要远好于Java相同的实现。  
