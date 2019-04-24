#!/bin/sh
curl -X POST \
  http://localhost/v1/Register \
  -F userId=farerpath \
  -F email=farerpath@fp.com \
  -F 'password=Test1234$'
curl -X POST \
  http://localhost/v1/Login \
  -F loginUserId=farerpath \
  -F 'loginPassword=Test1234$' \
  -F loginDeviceType=0
curl -X GET \
  http://localhost/v1/Album/farerpath
