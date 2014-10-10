RestTest
========

Quick and dirty way to test your API(rest)

./rtest filename=urllist.txt -Dapi_placeholder=api.john.pc



File format
===========
http-method | url | expected http return code | [optional text in response body]

get|http://myrestapi.com/user?id=123|200

OR

get|http://myrestapi.com/user?id=123|200|123

