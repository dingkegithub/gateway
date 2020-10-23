>
> [1. Cookie](#1)
>
> > [1.1 cookie 的工作原理](#1.1)
> >
> > [1.2 cookie 关键](#1.2)
>
> > [1.3 Set-Cookie 格式](#1.3)
>
> [2. jwt](#2)
>
> > [2.1 jwt 结构](#1.1)
>
> > 
>

---

<h2 id='1'> 1. Cookie </h2>

<h4 id='1.1'> 1.1 cookie 的工作原理 </h4>

<br>用户登录时，若登录成功服务器则会通过 Set-Cookie 设置登录信息，返回之后浏览器

会从响应头的 Set-Cookie 中取出这些信息保存到浏览器，然后下次用户在访问该服务时，

一旦访问路径和域满足 Set-Cookie 的设定，浏览器会在用户的请求头部添加 

Cookie: name=value; name=value..., 服务器收到请求后会从请求头中取出这些字段，用

来验证用户是否登录，获取用户信息，或是查看用户登录状态是否有效，或是判断Cookie

和请求中的用户一致，从而决定用户是否继续访问资源


<h4 id='1.2'> 1.2 cookie 关键 </h4>

    - 服务器通过 Set-Cookie 设置cookie

    ```
    HTTP/1.0 200 OK
    Content-type: text/html
    Set-Cookie: yummy_cookie=choco
    Set-Cookie: tasty_cookie=strawberry
    ```

    - 浏览器保存 Set-Cookie 中的键值对

    - 符合的请求浏览器会自动在请求头添加存储在浏览器中的 Cookie

    ```
    GET /sample_page.html HTTP/1.1
    Host: www.example.org
    Cookie: yummy_cookie=choco; tasty_cookie=strawberry
    ```

<h4 id='1.3'> 1.3 Set-Cookie 格式 </h4>


> Set-Cookie: key=value; Domain="xxx', Path='/', Max-Age=秒,Expire=日期; Secure; HttpOnly

    - **`key=value:`** 实际要设置的键值对

    - **`Domain='xxx':`** 指定接收cookie的主机, 不指定默认为 Origin 字段，子域名中不会发送cookie，若指定主域名，则子域名都带cookie

    - **`Path='/':`** 那些uri可接收cookie，/ 就是表示所有了，/api 表示凡是api开头的接收cookie

    - **`Expire=date:`** 设置过期日期，一般操作上就是登陆时的日期里加上一段时间，比如7天

    - **`Max-Age=秒:`** 0或-1直接过期，表示需要经过多少秒过期，优先级高于 **`Expire`**

    - **`Secure:`** 只有`https` 中发送cookie

    - **`HttpOnly:`** 阻止 Document.cookie 属性、XMLHttpRequest 和  Request APIs 访问cookie，以防范跨站脚本攻击（XSS）

<h2 id='2'> 2. jwt </h2>

<h4 id='2.1'> 2.1 jwt 结构 </h4>

    - **`header:`** 主要定义加密算法, `header` 最后会base64机密组成jwt的第一部分

        - **`alg:`** 算法定义, 比如 'HS256'
        - **`typ:`** jwt 类型, 一般固定为 'JWT'
        - **`cty:`** 内容类型
 
    - **`payload:`** 实际数据，标准声明部分和用户声明组成，用户声明就是键值对, `payload` 通过base64 加密得到第二部分
        
        - **`iss:`** 业务场景相关的一个字符串或是URI，若是URI可以代表jwt承载相关URI的数据，可以理解为jwt 的签发者
        - **`sub:`** 业务场景相关的一个字符串或是URI
        - **`aud:`** 接收 JWT 的接受者
        - **`exp:`** jwt 的过期时间
        - **`nbf:`** 和exp相反，表示在这个时间前jwt是无效的
        - **`jat:`** jwt 的签发时间
        - **`jti:`** jwt 的唯一标识, 即 JWT ID， 代表jwt的唯一性，

    - **`signature:`** 签名信息，algo(base64(header).base64(payload), salt-scret) = signature
    
        salt-secret 是非常重要，这部分一旦泄露，真个jwt字符串相当于是透明的

> jwt: base64(header).base64(payload).signature
