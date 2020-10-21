var express = require('express');

/*
 * origin: null
 * --------------------------------------------------------------------------------------------------------------------
 *  A null origin value typically indicates that the request is coming from a file on a user’s computer, rather than 
 *  from a website. It can also mean the request came from a redirect
 *   If you’re using the whitelist method, be sure to add the null value to your whitelist
 */
var originWhitelist = [
    'null',
    'http://localhost:1111'
];

var corsOptions = {
    allowOrigin: createWhitelistValidator(originWhitelist)
};

/*
 *
 * the Host header doesn’t include any scheme information, so it’s up to you to decide
 * which scheme to use
*/
var isSameOrigin = function(req) {
    var host = req.protocol + '://' + req.headers['host'];
    var origin = req.headers['origin'];
    return host === origin || !origin;
};

/*
 * 1 Grab the value from the Origin header.
 * 2 Validate the origin value using your chosen technique.
 * 3 If the origin is valid, set the Access-Control-Allow-Origin header.
 *
 *
 * Vary: Origin
 * -----------------------------------------------------------------------------------------------------
 * a request from the origin http://localhost:1111 will return the header Access-Control-Allow-Origin:
 * http://localhost:1111, but a request from http://localhost:2222 to the same server will return the 
 * header Access-Control-Allow-Origin: http://localhost:2222. These different response headers from 
 * the same servers can sometimes cause caching issues. If your server can return different 
 * Access-Control-Allow-Origin headers to different clients, you should also set the Vary HTTP response 
 * header to Origin
 * Without the Vary header, proxy servers may cache responses for one client and send them as responses 
 * to a different client.
 *
 * Bob is using his iPhone and visits http://mobile.espn.com, while Alice is using her tablet and visits 
 * http://tablet.espn.com. Because both Alice and Bob are at work, their requests flow through the company’s 
 * proxy server.  When Alice makes the first request to http://tablet.espn.com, the tablet site makes
 * a CORS request to load the scores from http://api.espn.com. The API responds with the header 
 * Access-Control-Allow-Origin: http://tablet.espn.com, and the proxy server caches the response.
 * Next, Bob makes his request to http://mobile.espn.com, and the mobile site grabs the scores from the same API. 
 * The proxy server notices that the request is to the same server that the tablet requested, and so it returns 
 * the cached response. Unfortunately, the cached response has the Access-Control-Allow-Origin: http://tablet.espn.com
 * header set. This header causes a request from http://mobile.espn.com to fail, because the Origin header doesn’t 
 * match the Access-Control-Allow-Origin header (figure 6.6).  Luckily, there is a way to fix this. The Vary header
 * tells the proxy server that the Origin header should be taken into account when deciding whether or not to send
 * cached content. With the Vary: Origin header in place, the proxy server will treat a request with Origin: http://mobile.espn.com
 * differently from a request with Origin: http://tablet.espn.com.
 *
 *
 */
var createWhitelistValidator = function(whitelist) {
    return function(val) {
        for (var i = 0; i < whitelist.length; i++) {
            if (val === whitelist[i]) {
                return true;
            }
        }
        return false;
    }
};

var handleCors = function(options) { 
    return function(req, res, next) {
        if (options.allowOrigin) {
            var origin = req.headers['origin'];
            if (options.allowOrigin(origin)) {
                res.set('Access-Control-Allow-Origin', origin);
            }
            res.set('Vary', 'Origin');
        } else {
            res.set('Access-Control-Allow-Origin', '*');
        }

        res.set('Access-Control-Allow-Origin', 'http://localhost:1111');
        res.set('Access-Control-Allow-Credentials', 'true');

        if (isPreflight(req)) {
            res.set('Access-Control-Allow-Methods', 'GET', 'DELETE');
            res.set('Access-Control-Allow-Headers',
                'Timezone-Offset, Sample-Source');
            res.set('Access-Control-Max-Age', '120');
            res.status(204).end();
            return;
        } else {
            res.set('Access-Control-Expose-Headers', 'X-Powered-By');
        }
        next();
    }
};

var SERVER_PORT = 9999;
var serverapp = express();

serverapp.use(cookieParser());
serverapp.use(express.static (__dirname));
serverapp.use(handleCors(corsOptions)); 

serverapp.get('/api/posts', function(req, res) {
    res.json(POSTS);
});

/*
### Cors

 cors: cross origin resource sharding
if any condition the folowing not match will lead to cors

1. schema must equal
2. host string must equal
3. port must equal

### preflight

following sitituation does not send preflight:

1. It uses an HTTP method other than GET, POST, or HEAD
2. Content-Type
   application/x-www-form-urlencoded
   multipart/form-data
   text/plain

3. header
   Accept
   Accept-Language
    Content-Language

4. The XMLHttpRequest contains upload events


###  isPreflight method checks three things:

   - Is the request an HTTP OPTIONS request?
   - Does the request have an Origin header?
   - Does the request have an Access-Control-Request-Method header?

### method permision

	- client: Access-Control-Request-Method: x
	- server: Access-Control-Allow-Methods: x,x,x

### header permision

	- client: Access-Control-Request-Headers: x,x,x
	- server: Access-Control-Allow-Headers: x,x,x,x

### how to response preflight:

    - set Access-Control-Allow-Origin
	- set Access-Control-Allow-Methods
	- if supply Access-Control-Allow-Headers set Access-Control-Allow-Headers

### preflight cache of browser

    Access-Control-Max-Age:  how long a preflight response is cached through
the Access-Control-Max-Age response header, browser cache expire

### cookie

	- server side:
	AccessControl-Allow-Credentials: 跨域请求中是否允许cookie,
	 If the request includes a preflight request, the Access-Control-Allow-Credentials
header must be present on both the preflight and the actual request. But the cookie
will only be sent on the actual request; the preflight request will never have a cookie.

	- client side:
	  withCredentials: 跨域请求中是否可以包含cookie
	     var xhr = createXhr('DELETE', url);
		 xhr.withCredentials = true; // Indicates that cookies are included with the request

###  How withCredentials and Access-Control-Allow-Credentials interact

|withCredentials |   Access-Control-Allow-Credentials | status |
|---|---|---|
|false|                 false|                         Allowed, but cookie include in the request|
|true |                 true |                         Allowed, cookie include in the request|
|false|                 true |                         Allowed, but cookie include in the request|
|true |                 false|                         Rejected,  Invalid because cookies are sent on the request, but the server doesn’t allow them|

ex1:

- client

```
   var xhr = new XMLHttpRequest();
   xhr.open('GET','http://127.0.0.1:9999/api/posts');
   xhr.send();

----response-----
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:1111
```

ex2:

- cliennt: withCredentials: true
- server: Access-Control-Allow-Credentials: true

```
var xhr = new XMLHttpRequest();
xhr.open('GET', 'http://127.0.0.1:9999/api/posts');
xhr.withCredentials = true;
xhr.send();

----response-----
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:1111
Access-Control-Allow-Credentials: true
```

ex3:

- cliennt: withCredentials: false
- server: Access-Control-Allow-Credentials: true

 the server sets the Access-Control-Allow-Credentials header to true, even
though the client doesn’t set the withCredentials property. Although the values of
the two don’t match, the request still succeeds.

```
var xhr = new XMLHttpRequest();
xhr.open('GET', 'http://127.0.0.1:9999/api/posts');
xhr.send();

----response-----
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:1111
Access-Control-Allow-Credentials:true
```

ex4:

- cliennt: withCredentials: true
- server: Access-Control-Allow-Credentials: false

```
var xhr = new XMLHttpRequest();
xhr.open('GET', 'http://127.0.0.1:9999/api/posts');
xhr.withCredentials = true;
xhr.send();

----response-----
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:1111
```
*/
