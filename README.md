RestTest
========

Quick and dirty way to test your API(rest)

./rtest filename=url_list.txt -DAPI_PLACEHOLDER=api.john.pc



**File format**

    http-method | url | expected http return code | [optional text in response body]


**Check for 200 return code**

    get|http://API_PLACEHOLDER/user?id=123|200


**Check for '123' in the response as well as a 200 status code**

    get|http://API_PLACEHOLDER/user?id=123|200|123



**Check for a 200 return code - no string-replace for api-host**

    get|http://api.example.com/user?id=123|200|123


**Does a post and checks for a 201 status code and response text 'success'**

    post|http://api.example.com/user?name=koos&surname=tonder|201|success
