RestTest
========

Quick and dirty way to test your API(rest)

./rtest filename=url_list.txt -DAPI_PLACEHOLDER=api.john.pc



File format
===========
http-method | url | expected http return code | [optional text in response body]

# Checks for 200 return code
get|http://API_PLACEHOLDER/user?id=123|200

OR

# Check for 123 in response as well as 200 status code
get|http://API_PLACEHOLDER/user?id=123|200|123


OR

# Checks for 200 return code - no string-replace for api-host
get|http://api.example.com/user?id=123|200|123


# Does a post and checks for 201 and response text "success"
post|http://api.example.com/user?name=koos&surname=tonder|201|success







