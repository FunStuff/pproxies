# pproxies
代理之代理。

## install

Docker:

    docker build -t xlaurent/pproxies https://github.com/xlaurent/pproxies.git#master:pproxies
    docker run -d -p 9527:9527 -e TEST_URL="http://baidu.com" xlaurent/pproxies

go get:

    go get github.com/xlaurent/pproxies/pproxies

从release中下载

## usage

    pproxies [OPTIONS]

    Application Options:
      -p, --port=      listened port (default: 9527)
      -s, --chunksize= chunksize (default: 10)
      -u, --url=       test url (default: http://httpbin.org/ip)
      -T, --ttimeout=  test timeout(second) (default: 5)
      -t, --ptimeout=  proxy timeout(second) (default: 5)
      -e, --error=     max error (default: 30)
      -n, --num=       proxy number (default: 5)
          --api=       proxy api (default: default)
          --interval=  auto check interval(second) if 0,disable auto check
      --banstr=    if auto check find this string in response,it will switch proxy
    -c, --config=    config file

    Help Options:
    -h, --help       Show this help message

error: 上游代理允许的最大错误数。当超过时，自动更换上游代理。

num: 同时使用的上游代理数。

interval: 自动检测发起请求的间隔,如果是0，关闭自动检测

banstr: 若自动检测发现响应包含该字符串，则判定代理不可用，切换代理（仅当自动检测开启时有效）。

api: 获取代理的API，返回的格式为：
        
    127.0.0.1:8080
    192.168.1.1:9999
    ...
    
若不指定，则使用内置代理池
## 这是什么？

这是一个 HTTP 代理,支持 CONNECT 方法。

## 与其他HTTP代理，它有何不同之处？
    
pproxies 还有一个上游代理。上游代理从 pproxies 所维护的代理池中获取。**一旦目前的上游代理不可用，便会自动切换上游代理，从而与目的网站保持连接畅通。**

一般 HTTP 代理：HTTP 请求 -> 代理所在主机 -> 目的主机

pproxies：HTTP 请求 -> 代理所在主机 -> 上游代理主机 -> 目的主机

## 代理池中的代理是从哪里来的？

 1. 使用者提供能获取代理 IP 的 API 。
 2. 使用内置的代理 IP（从某网站获取的免费代理）。
 
当代理池为空时，会自动调用API获取新的代理。

## 它是如何判断代理的可用性？

代理的可用性是相对于网站而言的，因此，用户需要提供一个测试网址和超时时间。pproxies 会在以下两个地方判断代理的可用性：

1. 添加进代理池前：使用待测试代理连接测试网址，若超时或失败，则代理不可用，抛弃之。
2. 作为上游代理时：超过一定的错误数，则代理不可用，切换之。错误指的是：连接失败，或状态码为4XX,5XX。

## 什么是*自动检测*功能？

自动检测：pproxies会每隔一段时间主动向测试网址发出请求（使用上游代理），如果失败，状态码为4XX,5XX，或**响应包含特定字符**，则判断代理不可用，切换之。

自动检测功能主要针对两种情况：

1. HTTPS: 代理作为中间人，无法获取响应的内容，则无法得知代理是否被禁。而*自动检测*的请求者是 pproixes，可以读取响应，自然就能判断了。
2. 那些即使失败也返回状态码200的网站：*自动检测*会读取响应，若响应包含特定字符（由用户指定，如“请输入验证码”之类），则判断代理不可用，切换之。

