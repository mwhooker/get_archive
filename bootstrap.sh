#!/bin/bash

environment="development"
s3_url="http://wercker-${environment}"
head_location="${s3_url}/get_archive/master/HEAD"
HEAD=$(/usr/bin/curl $head_location)
app_location="${s3_url}/get_archive/master/${HEAD}/linux_amd64/build"

/usr/bin/curl -o /tmp/get_kp $app_location
/usr/bin/chmod 0544 /tmp/get_kp
